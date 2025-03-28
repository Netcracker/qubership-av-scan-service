package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/netcracker/qubership-av-scan-service/pkg/clamav"
	"github.com/netcracker/qubership-av-scan-service/pkg/errors"
	"github.com/netcracker/qubership-av-scan-service/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ScanStatus is a struct representing a single file scan status.
type ScanStatus struct {
	// Filename is the name of the file which was scanned.
	Filename string `json:"filename"`
	// Infected is true if virus was found
	Infected bool `json:"infected"`
	// Virus is a string representing found virus, set only if Infected
	Virus string `json:"virus,omitempty"`
}

// VirusesFoundMetric is the name of the metric which tracks
// total number of found viruses
const VirusesFoundMetric = "av_viruses_found_total"

// ScanHandler handles scan requests.
// It parses multipart/form-data to files and verifies each file on the fly.
type ScanHandler struct {
	clamd        clamav.Clamd
	virusesCount prometheus.Counter
}

func NewScanHandler(clamd clamav.Clamd, reg *prometheus.Registry) *ScanHandler {
	virusesCount := promauto.With(reg).NewCounter(
		prometheus.CounterOpts{Name: VirusesFoundMetric},
	)
	return &ScanHandler{clamd: clamd, virusesCount: virusesCount}
}

func (s *ScanHandler) Handle(req *http.Request) (any, error) {
	contentType := req.Header.Get("Content-Type")
	if !strings.Contains(contentType, "multipart/form-data") {
		return nil, errors.ContentTypeUnsupportedError(contentType)
	}

	scans := make([]*ScanStatus, 0)
	reader, err := req.MultipartReader()
	if err != nil {
		return nil, errors.RequestBodyReadError(err)
	}

	part, partErr := reader.NextPart()
	for partErr != io.EOF {
		if partErr != nil {
			return nil, errors.RequestBodyReadError(partErr)
		}

		filename := part.FileName()
		if filename == "" {
			return nil, errors.FilenameNotSpecifiedError()
		}

		res, err := s.clamd.ScanStream(req.Context(), part)
		if err != nil {
			return nil, errors.ClamdScanError(err)
		}

		if res.Infected {
			log.From(req).Warn(
				"virus detected",
				"virus", res.VirusDescription,
				"filename", filename,
			)
			s.virusesCount.Inc()
		}
		scans = append(scans, &ScanStatus{
			Filename: filename,
			Infected: res.Infected,
			Virus:    res.VirusDescription,
		})

		part, partErr = reader.NextPart()
	}

	return scans, nil
}

// ParseScanStatuses is used to decode JSON input to list of scan statuses
func ParseScanStatuses(r io.Reader) ([]*ScanStatus, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var scanStatuses []*ScanStatus
	err = json.Unmarshal(data, &scanStatuses)
	return scanStatuses, err
}
