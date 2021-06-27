package traceandtracego_test

import (
	"io/ioutil"
	"log"
	"net/http"

	tracing "github.com/codeandcode0x/traceandtrace-go"
)

func ExampleAddHttpTracing() {
	httpClient := &http.Client{}
	r, _ := http.NewRequest("GET", "http://www.weather.com.cn/data/sk/101010100.html", nil)
	// set tracing
	_, cancel := tracing.AddHttpTracing("HttpClent", r.Header, map[string]string{"version": "v1"})
	defer cancel()
	// send reqeust
	response, _ := httpClient.Do(r)
	if response.StatusCode == 200 {
		str, _ := ioutil.ReadAll(response.Body)
		bodystr := string(str)
		log.Println("body string", bodystr)
	}
}
