package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

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

func MakeProxiedHandler(proxyBaseUri string, logHandler log.LogHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// step [1] in here, we have access to proxyBaseUri because it is in our
		// parent scope

		// step [2], read body
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logHandler.Handle(w, err, http.StatusBadGateway)
			return
		}

		// step [3] create http request
		newRequestUri := fmt.Sprintf("%v%v%v", r.URL.Scheme, proxyBaseUri, r.URL.Path)
		newRequest, err := http.NewRequest(r.Method, newRequestUri, bytes.NewBuffer(bodyBytes))
		if err != nil {
			logHandler.Handle(w, err, http.StatusBadGateway)
			return
		}

		// step [4] copy headers onto new request
		for headerName, headerValueSlice := range r.Header {
			for _, headerValue := range headerValueSlice {
				newRequest.Header.Set(headerName, headerValue)
			}
		}

		// step [5] make a new client
		client := http.Client{}

		// step [6] perform request
		resp, err := client.Do(newRequest)
		if err != nil {
			logHandler.Handle(w, err, http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// step [7] copy response
		copyResponse(w, resp)
	}
}

func copyHeaders(w http.ResponseWriter, resp *http.Response) {
	for headerName, headerStringSlice := range resp.Header {
		for _, headerValue := range headerStringSlice {
			//TODO log details
			//logger.Info("response header: " + headerName + ": " + headerValue)
			w.Header().Set(headerName, headerValue)
		}
	}
}

func copyBody(w http.ResponseWriter, resp *http.Response) {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//TODO log details
		//logger.Error("error reading seomtrans response body: " + err.Error())
		return
	}
	//TODO log details
	//logger.Info("body: " + string(bodyBytes))
	w.Write(bodyBytes)
}

func copyStatusCode(w http.ResponseWriter, resp *http.Response) {
	//TODO log details
	//logger.Info(fmt.Sprintf("status code: %v", resp.StatusCode))
	w.WriteHeader(resp.StatusCode)
}

func copyResponse(w http.ResponseWriter, resp *http.Response) {
	copyHeaders(w, resp)
	copyStatusCode(w, resp)
	copyBody(w, resp)
}
