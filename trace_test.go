package traceandtracego

import (
	"net/http"

	tracing "github.com/codeandcode0x/traceandtrace-go"
)

func TraceReporterTest() {

	_, cancel := tracing.AddHttpTracing("test_http", http.Header{}, map[string]string{"version": "v1"})
	defer cancel()

}
