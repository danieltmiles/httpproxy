package proxy

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/monsooncommerce/log"
)

/*
 * 1. find URL to proxy to
 * 2. read original request body
 * 2.1 the reason for this is so we can inspect it but also have something to pass on
 * 3. create new HTTP request which includes a bytes.NewBuffer(bodyBytes)
 * 4. Copy headers onto new request
 * 5. make a new client (research if this is necessary)
 * 6. perform request
 * 7. copy response
 * 7.1 copy headers
 * 7.2 copy status code
 * 7.3 copy body
 */

func MakeProxiedHandler(proxyBaseUri string, logHandler log.LogHandler, logger *log.Log) http.HandlerFunc {
	// step [5] make a new client
	var client http.Client
	if os.Getenv("DOCKER_CONTAINER") == "" {
		client = http.Client{}
	} else {
		pemCerts, err := ioutil.ReadFile("/certs/ca-bundle.crt")
		if err != nil {
			logger.Error("Unable to read /certs/ca-bundle.crt (%v), HTTPS calls likely will not work", err)
			client = http.Client{}
		} else {
			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM(pemCerts)
			client = http.Client{
				Transport: &http.Transport{
					TLSClientConfig:     &tls.Config{RootCAs: pool},
					MaxIdleConnsPerHost: 90,
				},
			}
		}

	}
	return func(w http.ResponseWriter, r *http.Request) {
		// step [1] in here, we have access to proxyBaseUri because it is in our
		// parent scope

		// step [2], read body
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Errorf("unable to read incoming request body: %v", err)
			logHandler.Handle(w, err, http.StatusBadGateway)
			return
		}
		defer r.Body.Close()

		// step [3] create http request
		newRequestUri := fmt.Sprintf("%v%v%v", r.URL.Scheme, proxyBaseUri, r.URL.Path)
		newRequest, err := http.NewRequest(r.Method, newRequestUri, bytes.NewBuffer(bodyBytes))
		if err != nil {
			logger.Errorf("unable to form new HTTP request: %v", err)
			logHandler.Handle(w, err, http.StatusBadGateway)
			return
		}

		// step [4] copy headers onto new request
		for headerName, headerValueSlice := range r.Header {
			for _, headerValue := range headerValueSlice {
				newRequest.Header.Set(headerName, headerValue)
			}
		}

		// step [6] perform request
		resp, err := client.Do(newRequest)
		if err != nil {
			logger.Errorf("error from client.Do trying to reach %v (%v), %+v",
				newRequestUri, err, client.Transport)
			logHandler.Handle(w, err, http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// step [7] copy response
		copyResponse(w, resp, logger)
	}
}

func copyHeaders(w http.ResponseWriter, resp *http.Response) {
	for headerName, headerStringSlice := range resp.Header {
		for _, headerValue := range headerStringSlice {
			w.Header().Set(headerName, headerValue)
		}
	}
}

func copyBody(w http.ResponseWriter, resp *http.Response, logger *log.Log) {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Unable to read foreign response body: %v", err)
		return
	}
	defer resp.Body.Close()
	w.Write(bodyBytes)
}

func copyStatusCode(w http.ResponseWriter, resp *http.Response) {
	w.WriteHeader(resp.StatusCode)
}

func copyResponse(w http.ResponseWriter, resp *http.Response, logger *log.Log) {
	copyHeaders(w, resp)
	copyStatusCode(w, resp)
	copyBody(w, resp, logger)
}
