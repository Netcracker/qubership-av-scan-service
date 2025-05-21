package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"syscall"

	"github.com/netcracker/qubership-av-scan-service/pkg/clamav"
	"github.com/netcracker/qubership-av-scan-service/pkg/router"
	certwatcher "github.com/netcracker/qubership-av-scan-service/pkg/tls"

	"github.com/oklog/run"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "av-scan-service",
	Short: "av-scan-service",
	Long:  "Antivirus Scan Service allows to scan files for viruses using HTTP API",
	Run:   Run,
}

func init() {
	rootCmd.PersistentFlags().String("certfile", "", "SSL certificate file name")
	rootCmd.PersistentFlags().String("keyfile", "", "SSL key file name")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error executing command: %s", err)
		os.Exit(1)
	}
}

// CheckFile is used to check, if file is defined and exists
func CheckFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file is undefined")
	}
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("can't get info for %s: %s", filePath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s is not a file", filePath)
	}
	return nil
}

// ParseCertsFromArgs parses keyfile and certfile cli arguments, verifies them and returns
func ParseCertsFromArgs(cmd *cobra.Command, logger *slog.Logger) (string, string) {
	certFile, err := cmd.Flags().GetString("certfile")
	if err != nil {
		logger.Error("failed to get cert file", "error", err)
		os.Exit(1)
	}

	keyFile, err := cmd.Flags().GetString("keyfile")
	if err != nil {
		logger.Error("failed to get key file", "error", err)
		os.Exit(1)
	}

	if certFile != "" && keyFile != "" {
		if err := CheckFile(certFile); err != nil {
			logger.Error("failed to get cert file", "error", err)
			os.Exit(1)
		}
		if err := CheckFile(keyFile); err != nil {
			logger.Error("failed to get key file", "error", err)
			os.Exit(1)
		}
	} else if certFile != "" || keyFile != "" {
		logger.Error("cert file and key file should be specified together")
		os.Exit(1)
	}
	return certFile, keyFile
}

// ShutdownServer i,plement graceful shutdown for http server
func ShutdownServer(srv *http.Server, logger *slog.Logger, reason error) {
	logger.Warn("stopping antivirus scanning service", "reason", reason)
	err := srv.Shutdown(context.Background())
	if err != nil {
		logger.Error("failed to gracefully stop antivirus service", "error", err)
		os.Exit(1)
	}
	logger.Info("stopped antivirus scanning service")
}

func Run(cmd *cobra.Command, args []string) {
	level := &slog.LevelVar{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)
	logger.Info("started antivirus scanning service")

	certFile, keyFile := ParseCertsFromArgs(cmd, logger)
	tlsEnabled := certFile != ""

	// run http server
	r := router.NewRouter(clamav.NewClamD(), logger)
	var gr run.Group
	gr.Add(run.SignalHandler(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM))
	if tlsEnabled {
		watcher, err := certwatcher.New(certFile, keyFile, logger)
		if err != nil {
			logger.Error("error creating certificate watcher", "error", err)
			os.Exit(1)
		}
		srv := &http.Server{Handler: r, Addr: "0.0.0.0:8443"}
		srv.TLSConfig = &tls.Config{
			GetCertificate: watcher.GetCertificate,
		}
		gr.Add(func() error {
			return watcher.Watch()
		}, func(err error) {
			watcher.Stop()
		})

		gr.Add(func() error {
			err := srv.ListenAndServeTLS("", "")
			if err == http.ErrServerClosed {
				logger.Warn("http server closed")
			} else {
				logger.Error("failed to serve http", "error", err)
				os.Exit(1)
			}
			return err
		}, func(err error) {
			ShutdownServer(srv, logger, err)
		})
	} else {
		srv := &http.Server{Handler: r, Addr: "0.0.0.0:8080"}
		gr.Add(func() error {
			err := srv.ListenAndServe()
			if err == http.ErrServerClosed {
				logger.Warn("http server closed")
			} else {
				logger.Error("failed to serve http", "error", err)
				os.Exit(1)
			}
			return err
		}, func(err error) {
			ShutdownServer(srv, logger, err)
		})
	}

	if err := gr.Run(); err != nil {
		logger.Info("terminating...", "reason", err)
	}
}
