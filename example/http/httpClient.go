package main

import (
	"io/ioutil"
	"log"
	"net/http"

	tracing "github.com/codeandcode0x/traceandtrace-go"
)

func main() {
	httpClient()
}

// http tracing example
func httpClient() {
	// http server url
	httpTogRPCSrcUrl := "http://localhost:9090/rpc/tracing"
	// http request
	httpClient := &http.Client{}
	r, _ := http.NewRequest("GET", httpTogRPCSrcUrl, nil)
	// set tracing
	_, cancel := tracing.AddHttpTracing("HttpClent", "rpc/tracing GET", r.Header, map[string]string{"version": "v1"})
	defer cancel()
	// send reqeust
	response, _ := httpClient.Do(r)
	if response.StatusCode == 200 {
		str, _ := ioutil.ReadAll(response.Body)
		bodystr := string(str)
		log.Println("body string", bodystr)
	}
}
