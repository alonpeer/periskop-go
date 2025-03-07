package periskop

import (
	"errors"
	"net/http"
	"testing"
)

func getFirstAggregatedErr(aggregatedErrors map[string]*aggregatedError) *aggregatedError {
	for _, value := range aggregatedErrors {
		return value
	}
	return nil
}

func TestCollector_addError(t *testing.T) {
	c := NewErrorCollector()
	err := errors.New("testing")
	c.addError(err, SeverityError, nil, "")

	if len(c.aggregatedErrors) != 1 {
		t.Errorf("expected one element")
	}

	c.addError(err, SeverityError, nil, "")
	if getFirstAggregatedErr(c.aggregatedErrors).TotalCount != 2 {
		t.Errorf("expected two elements")
	}
}

func TestCollector_ReportError(t *testing.T) {
	c := NewErrorCollector()
	err := errors.New("testing")
	c.ReportError(err)

	if len(c.aggregatedErrors) != 1 {
		t.Errorf("expected one element")
	}

	errorWithContext := getFirstAggregatedErr(c.aggregatedErrors).LatestErrors[0]
	if errorWithContext.Error.Message != err.Error() {
		t.Errorf("expected a propagated error")
	}

	if errorWithContext.Error.Class != "*errors.errorString" {
		t.Errorf("incorrect class name, got %s", errorWithContext.Error.Class)
	}

	if len(errorWithContext.Error.Stacktrace) == 0 {
		t.Errorf("expected a collected stack trace")
	}

	if errorWithContext.Severity != SeverityError {
		t.Errorf("incorrect severity, got %s", SeverityError)
	}
}

func TestErrorCollector_ReportWithSeverity(t *testing.T) {
	c := NewErrorCollector()
	err := errors.New("testing")
	c.ReportWithSeverity(err, SeverityInfo)

	if len(c.aggregatedErrors) != 1 {
		t.Errorf("expected one element")
	}

	errorWithContext := getFirstAggregatedErr(c.aggregatedErrors).LatestErrors[0]
	if errorWithContext.Error.Message != err.Error() {
		t.Errorf("expected a propagated error")
	}

	if errorWithContext.Error.Class != "*errors.errorString" {
		t.Errorf("incorrect class name, got %s", errorWithContext.Error.Class)
	}

	if len(errorWithContext.Error.Stacktrace) == 0 {
		t.Errorf("expected a collected stack trace")
	}

	if errorWithContext.Severity != SeverityInfo {
		t.Errorf("incorrect severity, got %s", errorWithContext.Severity)
	}
}

func TestCollector_Report_errKey(t *testing.T) {
	c := NewErrorCollector()
	err := errors.New("testing")
	errKey := "grouped-err"
	errClass := "*errors.errorString"
	c.Report(ErrorReport{err: err, errKey: errClass + "@" + errKey})

	if len(c.aggregatedErrors) != 1 {
		t.Errorf("expected one element")
	}
	aggregatedErr := getFirstAggregatedErr(c.aggregatedErrors)
	errorWithContext := aggregatedErr.LatestErrors[0]
	if errorWithContext.Error.Message != err.Error() {
		t.Errorf("expected a propagated error")
	}

	if aggregatedErr.AggregationKey != errClass+"@"+errKey {
		t.Errorf("expected an overwritten key")
	}

	if errorWithContext.Error.Class != errClass {
		t.Errorf("incorrect class name, got %s", errorWithContext.Error.Class)
	}

	if len(errorWithContext.Error.Stacktrace) == 0 {
		t.Errorf("expected a collected stack trace")
	}

	if errorWithContext.Severity != SeverityError {
		t.Errorf("incorrect severity, got %s", SeverityError)
	}
}

func TestCollector_ReportWithHTTPContext(t *testing.T) {
	c := NewErrorCollector()
	body := "some body"
	err := errors.New("testing")
	httpContext := HTTPContext{
		RequestMethod:  "GET",
		RequestURL:     "http://example.com",
		RequestHeaders: map[string]string{"Cache-Control": "no-cache"},
		RequestBody:    &body,
	}
	c.ReportWithHTTPContext(err, &httpContext)

	if len(c.aggregatedErrors) != 1 {
		t.Errorf("expected one element")
	}

	errorWithContext := getFirstAggregatedErr(c.aggregatedErrors).LatestErrors[0]
	if errorWithContext.HTTPContext.RequestMethod != "GET" {
		t.Errorf("expected HTTP method GET")
	}

	if errorWithContext.Error.Class != "*errors.errorString" {
		t.Errorf("incorrect class name, got %s", errorWithContext.Error.Class)
	}

	if errorWithContext.Severity != SeverityError {
		t.Errorf("incorrect severity, got %s", SeverityError)
	}
}

func TestErrorCollector_ReportWithHTTPContextAndSeverity(t *testing.T) {
	c := NewErrorCollector()
	body := "some body"
	err := errors.New("testing")
	httpContext := HTTPContext{
		RequestMethod:  "GET",
		RequestURL:     "http://example.com",
		RequestHeaders: map[string]string{"Cache-Control": "no-cache"},
		RequestBody:    &body,
	}
	c.ReportWithHTTPContextAndSeverity(err, SeverityWarning, &httpContext)

	if len(c.aggregatedErrors) != 1 {
		t.Errorf("expected one element")
	}

	errorWithContext := getFirstAggregatedErr(c.aggregatedErrors).LatestErrors[0]
	if errorWithContext.HTTPContext.RequestMethod != "GET" {
		t.Errorf("expected HTTP method GET")
	}

	if errorWithContext.Error.Class != "*errors.errorString" {
		t.Errorf("incorrect class name, got %s", errorWithContext.Error.Class)
	}

	if errorWithContext.Severity != SeverityWarning {
		t.Errorf("incorrect severity, got %s", errorWithContext.Severity)
	}
}

func TestErrorCollector_ReportWithHTTPRequestAndSeverity(t *testing.T) {
	c := NewErrorCollector()
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = errors.New("testing")
	c.ReportWithHTTPRequestAndSeverity(err, SeverityInfo, req)

	if len(c.aggregatedErrors) != 1 {
		t.Errorf("expected one element")
	}

	errorWithContext := getFirstAggregatedErr(c.aggregatedErrors).LatestErrors[0]
	if errorWithContext.HTTPContext.RequestMethod != "GET" {
		t.Errorf("expected HTTP method GET")
	}

	if errorWithContext.HTTPContext.RequestBody != nil {
		t.Errorf("expected nil http request body but got %s", *errorWithContext.HTTPContext.RequestBody)
	}

	if errorWithContext.Error.Class != "*errors.errorString" {
		t.Errorf("incorrect class name, got %s", errorWithContext.Error.Class)
	}

	if errorWithContext.Severity != SeverityInfo {
		t.Errorf("incorrect severity, got %s", errorWithContext.Severity)
	}
}

func TestCollector_ReportWithHTTPRequest(t *testing.T) {
	c := NewErrorCollector()
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = errors.New("testing")
	c.ReportWithHTTPRequest(err, req)

	if len(c.aggregatedErrors) != 1 {
		t.Errorf("expected one element")
	}

	errorWithContext := getFirstAggregatedErr(c.aggregatedErrors).LatestErrors[0]
	if errorWithContext.HTTPContext.RequestMethod != "GET" {
		t.Errorf("expected HTTP method GET")
	}

	if errorWithContext.HTTPContext.RequestBody != nil {
		t.Errorf("expected nil http request body but got %s", *errorWithContext.HTTPContext.RequestBody)
	}

	if errorWithContext.Error.Class != "*errors.errorString" {
		t.Errorf("incorrect class name, got %s", errorWithContext.Error.Class)
	}

	if errorWithContext.Severity != SeverityError {
		t.Errorf("incorrect severity, got %s", SeverityError)
	}
}

func TestCollector_ReportErrorWithContext(t *testing.T) {
	c := NewErrorCollector()
	body := "some body"
	httpContext := HTTPContext{
		RequestMethod:  "GET",
		RequestURL:     "http://example.com",
		RequestHeaders: map[string]string{"Cache-Control": "no-cache"},
		RequestBody:    &body,
	}
	errorInstance := NewCustomErrorInstance("testing", "manual_error", []string{"line 0:", "error in testingError"})
	errorWithContext := NewErrorWithContext(errorInstance, SeverityError, &httpContext)
	c.ReportErrorWithContext(errorWithContext, SeverityError, "")

	if len(c.aggregatedErrors) != 1 {
		t.Errorf("expected one element")
	}

	errorWithContext = getFirstAggregatedErr(c.aggregatedErrors).LatestErrors[0]
	if errorWithContext.HTTPContext.RequestMethod != "GET" {
		t.Errorf("expected HTTP method GET")
	}

	if errorWithContext.Error.Class != "manual_error" {
		t.Errorf("incorrect class name, got %s", errorWithContext.Error.Class)
	}

	if errorWithContext.Severity != SeverityError {
		t.Errorf("incorrect severity, got %s", SeverityError)
	}
}

func TestCollector_getAggregatedErrors(t *testing.T) {
	c := NewErrorCollector()
	err := errors.New("testing")
	c.addError(err, SeverityError, nil, "")

	aggregatedErr := getFirstAggregatedErr(c.aggregatedErrors)
	payload := c.getAggregatedErrors()
	if payload.AggregatedErrors[0].AggregationKey != aggregatedErr.AggregationKey {
		t.Errorf("keys for aggregated errors are different, expected: %s, got: %s",
			aggregatedErr.AggregationKey, payload.AggregatedErrors[0].AggregationKey)
	}
}

func TestCollector_getStackTrace(t *testing.T) {
	err := errors.New("testing")
	stacktrace := getStackTrace(err)
	if len(stacktrace) == 0 {
		t.Errorf("expected a  stacktrace")
	}
	lastFrame := stacktrace[len(stacktrace)-1]
	if lastFrame == "" {
		t.Errorf("got empty frame")
	}
}
