package clamav

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	clamdClient "github.com/dutchcoders/go-clamd"
)

// clamdDateFormat is a time format in which clamd outputs its DB age,
// this format is used for actual DB age parsing
const clamdDateFormat = "Jan _2 03:04:05 2006"

type ScanResult struct {
	// Infected is true when there is a virus found, false otherwise
	Infected bool
	// VirusDescription is set to found Virus Description if Infected is true
	VirusDescription string
}

// Clamd is an interface used to work with underlying Clamd instance
type Clamd interface {
	// ScanStream scans given stream for viruses
	ScanStream(ctx context.Context, r io.Reader) (ScanResult, error)
	// DatabaseAge returns clamav DB age in seconds
	DatabaseAge() (float64, error)
	// Ping checks that Clamd is alive
	Ping() error
}

// clamdImpl implements Clamd interface using real clamd instance
type clamdImpl struct {
	client *clamdClient.Clamd
}

func (c *clamdImpl) ScanStream(ctx context.Context, r io.Reader) (ScanResult, error) {
	abort := make(chan bool)
	defer close(abort)

	ch, err := c.client.ScanStream(r, abort)
	if err != nil {
		return ScanResult{}, fmt.Errorf("failed to scan stream: %s", err)
	}

	select {
	case res := <-ch:
		if res == nil {
			return ScanResult{}, fmt.Errorf("failed to get clamd scan result, client failed unexpectedly")
		} else if res.Status == clamdClient.RES_FOUND {
			return ScanResult{Infected: true, VirusDescription: res.Description}, nil
		} else if res.Status != clamdClient.RES_OK {
			return ScanResult{}, fmt.Errorf("unexpected scan result: %s", res.Raw)
		}
	case <-ctx.Done():
		return ScanResult{}, fmt.Errorf("context closed during scanning: %s", ctx.Err())
	}

	return ScanResult{}, nil
}

func (c *clamdImpl) Ping() error {
	return c.client.Ping()
}

func (c *clamdImpl) DatabaseAge() (float64, error) {
	resCh, err := c.client.Version()
	if err != nil {
		return 0, err
	}

	res := <-resCh
	if res == nil {
		return 0, fmt.Errorf("failed to get ClamAV version, client failed unexpectedly")
	}

	versionString := strings.ReplaceAll(res.Raw, "  ", " ")
	versionParts := strings.SplitAfterN(versionString, " ", 3)
	dateString := versionParts[len(versionParts)-1]
	date, err := time.Parse(clamdDateFormat, dateString)
	if err != nil {
		return 0, fmt.Errorf("failed to parse DB date \"%s\", error: %s", res.Raw, err)
	}

	return time.Since(date).Seconds(), nil
}

// NewClamD creates default Clamd
func NewClamD() Clamd {
	return &clamdImpl{client: clamdClient.NewClamd("tcp://127.0.0.1:3310")}
}
