package testutils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/netcracker/qubership-av-scan-service/pkg/clamav"
)

// EICARTest contains text of the EICAR test file that EICAR developed specifically
// to test the response of computer antivirus programs (instead of using real malware)
// https://en.wikipedia.org/wiki/EICAR_test_file
const EICARTest = "X5O!P%@AP[4\\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*"

type ClamdMock struct {
	unhealthyReason string
	virusSignatures []string
}

func (c *ClamdMock) ScanStream(_ context.Context, r io.Reader) (clamav.ScanResult, error) {
	if c.unhealthyReason != "" {
		return clamav.ScanResult{}, errors.New(c.unhealthyReason)
	}

	content, err := io.ReadAll(r)
	if err != nil {
		return clamav.ScanResult{}, fmt.Errorf("failed to read content: %s", err)
	}

	for _, signature := range c.virusSignatures {
		if strings.Contains(string(content), signature) {
			return clamav.ScanResult{Infected: true, VirusDescription: signature}, nil
		}
	}

	return clamav.ScanResult{}, nil
}

func (c *ClamdMock) Ping() error {
	if c.unhealthyReason != "" {
		return errors.New(c.unhealthyReason)
	}
	return nil
}

func (c *ClamdMock) DatabaseAge() (float64, error) {
	return 0, nil
}

func NewClamdMock() *ClamdMock {
	// by default include only EICARTest signature
	virusSignatures := []string{EICARTest}
	return &ClamdMock{virusSignatures: virusSignatures}
}

func (c *ClamdMock) WithUnhealthy(reason string) *ClamdMock {
	c.unhealthyReason = reason
	return c
}
