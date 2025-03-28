package router_test

import (
	"bytes"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/netcracker/qubership-av-scan-service/pkg/clamav"
	"github.com/netcracker/qubership-av-scan-service/pkg/errors"
	"github.com/netcracker/qubership-av-scan-service/pkg/handlers"
	"github.com/netcracker/qubership-av-scan-service/pkg/router"
	"github.com/netcracker/qubership-av-scan-service/pkg/testutils"
	"github.com/prometheus/common/expfmt"
)

func TestHealthOK(t *testing.T) {
	r := router.NewRouter(testutils.NewClamdMock(), nil)

	respWriter := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(respWriter, req)

	resp := respWriter.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 status, but got: %v", resp.Status)
	}
}

func TestHealthBad(t *testing.T) {
	unhealthyReason := "uNhEaLtHy"
	r := router.NewRouter(testutils.NewClamdMock().WithUnhealthy(unhealthyReason), nil)

	respWriter := httptest.NewRecorder()
	r.ServeHTTP(respWriter, httptest.NewRequest(http.MethodGet, "/health", nil))

	resp := respWriter.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 status, but got: %v", resp.Status)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("expected application/json content type, but got: %s", resp.Header.Get("Content-Type"))
	}

	apiError, err := errors.Parse(resp.Body)
	if err != nil {
		t.Fatalf("failed to get error from body: %s", err)
	}

	if apiError.Details != unhealthyReason {
		t.Fatalf("expected api error message to be '%s', but got: '%s'",
			unhealthyReason, apiError.Details)
	}
}

func TestMetricsPresent(t *testing.T) {
	// send scan request once for HTTP metrics to appear
	r := router.NewRouter(testutils.NewClamdMock(), slog.Default())
	r.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/api/v1/scan", strings.NewReader("test")),
	)

	// get metrics response
	respWriter := httptest.NewRecorder()
	r.ServeHTTP(respWriter, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	resp := respWriter.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected OK response, but got: %v", resp.Status)
	}

	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		t.Fatalf("expected to parse prometheus metrics, but failed: %s", err)
	}

	// verify required metrics present
	expectedMetrics := []string{
		"go_gc_duration_seconds",    // an example GC metric
		"process_cpu_seconds_total", // an example process metric
		"http_request_duration_seconds",
		"http_request_size_bytes",
		"http_requests_inflight",
		"http_requests_total",
		handlers.VirusesFoundMetric,
		clamav.DatabaseAgeMetric,
	}
	for _, metricName := range expectedMetrics {
		if _, ok := mf[metricName]; !ok {
			t.Fatalf("%s not found", metricName)
		}
	}
}

func TestScanNotInfected(t *testing.T) {
	r := router.NewRouter(testutils.NewClamdMock(), slog.Default())
	respWriter := httptest.NewRecorder()

	buffer := &bytes.Buffer{}
	multi := multipart.NewWriter(buffer)
	writeFile(multi, "file1", "safe content")
	multi.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scan", buffer)
	req.Header.Add("Content-Type", multi.FormDataContentType())
	r.ServeHTTP(respWriter, req)

	resp := respWriter.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected OK response, but got: %v", resp.Status)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("expected application/json content type, but got: %s", resp.Header.Get("Content-Type"))
	}

	statuses, err := handlers.ParseScanStatuses(resp.Body)
	if err != nil {
		t.Fatalf("expected to read scan statuses, but failed: %s", err)
	}

	if len(statuses) != 1 {
		t.Fatalf("expected excactly one status, but got: %d", len(statuses))
	}

	if statuses[0].Infected {
		t.Fatalf("expected infected to be false, but got true")
	}
}

func TestScanInfected(t *testing.T) {
	r := router.NewRouter(testutils.NewClamdMock(), slog.Default())
	respWriter := httptest.NewRecorder()

	buffer := &bytes.Buffer{}
	multi := multipart.NewWriter(buffer)
	writeFile(multi, "file1", testutils.EICARTest)
	multi.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scan", buffer)
	req.Header.Add("Content-Type", multi.FormDataContentType())
	r.ServeHTTP(respWriter, req)

	resp := respWriter.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected OK response, but got: %v", resp.Status)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("expected application/json content type, but got: %s", resp.Header.Get("Content-Type"))
	}

	statuses, err := handlers.ParseScanStatuses(resp.Body)
	if err != nil {
		t.Fatalf("expected to read scan statuses, but failed: %s", err)
	}

	if len(statuses) != 1 {
		t.Fatalf("expected excactly one status, but got: %d", len(statuses))
	}

	if !statuses[0].Infected {
		t.Fatalf("expected infected to be true, but got false")
	}
}

func TestVirusesCountMetricIncremented(t *testing.T) {
	// send scan request with test virus to increment counter
	r := router.NewRouter(testutils.NewClamdMock(), slog.Default())
	respWriter := httptest.NewRecorder()

	buffer := &bytes.Buffer{}
	multi := multipart.NewWriter(buffer)
	writeFile(multi, "file1", testutils.EICARTest)
	multi.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scan", buffer)
	req.Header.Add("Content-Type", multi.FormDataContentType())
	r.ServeHTTP(respWriter, req)

	// get metrics response
	respWriter = httptest.NewRecorder()
	r.ServeHTTP(respWriter, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	resp := respWriter.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected OK response, but got: %v", resp.Status)
	}

	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		t.Fatalf("expected to parse prometheus metrics, but failed: %s", err)
	}

	v, ok := mf[handlers.VirusesFoundMetric]
	if !ok {
		t.Fatalf("viruses_found_total metric not found")
	}
	if *v.Metric[0].Counter.Value != 1 {
		t.Fatalf("expected counter to increment to 1, but got: %f", *v.Metric[0].Counter.Value)
	}
}

func TestMultiScan(t *testing.T) {
	r := router.NewRouter(testutils.NewClamdMock(), slog.Default())
	respWriter := httptest.NewRecorder()

	buffer := &bytes.Buffer{}
	multi := multipart.NewWriter(buffer)
	writeFile(multi, "file1", "safe content")
	writeFile(multi, "file2", testutils.EICARTest)
	multi.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scan", buffer)
	req.Header.Add("Content-Type", multi.FormDataContentType())
	r.ServeHTTP(respWriter, req)

	resp := respWriter.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected OK response, but got: %v", resp.Status)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("expected application/json content type, but got: %s", resp.Header.Get("Content-Type"))
	}

	statuses, err := handlers.ParseScanStatuses(resp.Body)
	if err != nil {
		t.Fatalf("expected to read scan statuses, but failed: %s", err)
	}

	if len(statuses) != 2 {
		t.Fatalf("expected excactly two statuses, but got: %d", len(statuses))
	}

	for _, s := range statuses {
		if s.Filename == "file1" {
			if s.Infected {
				t.Fatalf("expected file1 to be not infected, but got infected")
			}
		} else if s.Filename == "file2" {
			if !s.Infected {
				t.Fatalf("expected file2 to be infected, but got non-infected")
			}
		} else {
			t.Fatalf("unexpected file name: %s", s.Filename)
		}
	}
}

func TestScanClamdProblem(t *testing.T) {
	unhealthyReason := "uNhEaLtHy"
	r := router.NewRouter(
		testutils.NewClamdMock().WithUnhealthy(unhealthyReason),
		slog.Default(),
	)
	respWriter := httptest.NewRecorder()

	buffer := &bytes.Buffer{}
	multi := multipart.NewWriter(buffer)
	writeFile(multi, "file1", "safe content")
	multi.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scan", buffer)
	req.Header.Add("Content-Type", multi.FormDataContentType())
	r.ServeHTTP(respWriter, req)

	resp := respWriter.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 response, but got: %v", resp.Status)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("expected application/json content type, but got: %s", resp.Header.Get("Content-Type"))
	}

	apiErr, err := errors.Parse(resp.Body)
	if err != nil {
		t.Fatalf("expected to read apiErr, but failed: %s", err)
	}

	if apiErr.Details != unhealthyReason {
		t.Fatalf("expected reason to be '%s', but got: %s",
			unhealthyReason, apiErr.Details)
	}
}

func TestScanFilenameNotSpecified(t *testing.T) {
	r := router.NewRouter(testutils.NewClamdMock(), slog.Default())
	respWriter := httptest.NewRecorder()

	buffer := &bytes.Buffer{}
	multi := multipart.NewWriter(buffer)
	writeFile(multi, "", "safe content")
	multi.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scan", buffer)
	req.Header.Add("Content-Type", multi.FormDataContentType())
	r.ServeHTTP(respWriter, req)

	resp := respWriter.Result()
	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415 response, but got: %v", resp.Status)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("expected application/json content type, but got: %s", resp.Header.Get("Content-Type"))
	}

	apiErr, err := errors.Parse(resp.Body)
	if err != nil {
		t.Fatalf("expected to read apiErr, but failed: %s", err)
	}

	if apiErr.Code != "AV-5001" {
		t.Fatalf("expected error code to be '%s', but got: %s",
			"AV-5001", apiErr.Code)
	}
}

func writeFile(w *multipart.Writer, name string, content string) {
	fileWrite, _ := w.CreateFormFile(name, name)
	_, err := fileWrite.Write([]byte(content))
	if err != nil {
		panic(err)
	}
}
