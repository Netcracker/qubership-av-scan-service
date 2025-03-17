package certwatcher

import (
	"crypto/tls"
	"log/slog"
	"path/filepath"
	"sync"
	"fmt"

	"github.com/fsnotify/fsnotify"
)

type CertWatcher struct {
	mu       sync.RWMutex
	certFile string
	keyFile  string
	keyPair  *tls.Certificate
	watcher  *fsnotify.Watcher
	watching chan bool
	logger   *slog.Logger
}

func New(certFile, keyFile string, logger *slog.Logger) (*CertWatcher, error) {
	var err error

	if logger == nil {
		logger = slog.Default()
	}

	certFile, err = filepath.Abs(certFile)
	if err != nil {
		return nil, err
	}

	keyFile, err = filepath.Abs(keyFile)
	if err != nil {
		return nil, err
	}

	cw := &CertWatcher{
		mu:       sync.RWMutex{},
		certFile: certFile,
		keyFile:  keyFile,
		logger:   logger,
	}

	return cw, nil
}

func (cw *CertWatcher) Watch() error {
	var err error

	if cw.watcher, err = fsnotify.NewWatcher(); err != nil {
		return fmt.Errorf("can't create watcher for certificates: %w", err)
	}

	if err = cw.watcher.Add(cw.certFile); err != nil {
		return fmt.Errorf("can't watch cert file: %w", err)
	}

	if err = cw.watcher.Add(cw.keyFile); err != nil {
		return fmt.Errorf("can't watch key file: %w", err)
	}

	if err := cw.load(); err != nil {
		cw.logger.Error("can't load cert or key file", "error", err)
		return err
	}

	cw.logger.Info("watching for cert and key change")

	cw.watching = make(chan bool)

	cw.run()
	return nil
}

func (cw *CertWatcher) load() error {
	keyPair, err := tls.LoadX509KeyPair(cw.certFile, cw.keyFile)
	if err == nil {
		cw.mu.Lock()
		cw.keyPair = &keyPair
		cw.mu.Unlock()
		cw.logger.Info("certificate and key loaded")
	}

	return err
}

func (cw *CertWatcher) run() {
loop:
	for {
		select {
		case <-cw.watching:
			break loop
		case event, ok := <-cw.watcher.Events:
			if !ok {
				break loop
			}
			cw.logger.Info("watch certificate event", "event", event)

			if event.Op.Has(fsnotify.Remove) || event.Op.Has(fsnotify.Chmod) {
				if err := cw.watcher.Add(event.Name); err != nil {
					cw.logger.Error("error re-watching file", "error", err)
				}
			}
			if err := cw.load(); err != nil {
				// Is new certificate is wrong, the old one will be used.
				// This approach is based on controller-runtime cert-watcher: 
				// https://github.com/kubernetes-sigs/controller-runtime/blob/v0.18.4/pkg/certwatcher/certwatcher.go#L189-L191
				cw.logger.Error("can't load cert or key file", "error", err)
			}
		case err, ok := <-cw.watcher.Errors:
			if !ok {
				break loop
			}
			cw.logger.Error("error watching certificates", "error", err)
		}
	}

	cw.logger.Info("stopped certificate watching")
	cw.watcher.Close()
}

func (cw *CertWatcher) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	return cw.keyPair, nil
}

func (cw *CertWatcher) Stop() {
	cw.watching <- false
}
