package proxy

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/franela/goblin"
	"github.com/monsooncommerce/log"
	"github.com/monsooncommerce/mockWriter"
	. "github.com/onsi/gomega"
)

func TestProxy(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Proxy test", func() {
		var logger *log.Log
		g.BeforeEach(func() {
			logger = log.New(mockwriter.New(), log.Debug)
		})
		g.It("should proxy a basic request", func() {
			proxyToThis := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write([]byte("these are the bytes from proxied service."))
			}))
			logHandler := log.LogHandlerImpl{logger}

			server := httptest.NewServer(MakeProxiedHandler(proxyToThis.URL, &logHandler))
			req, err := http.NewRequest("GET", server.URL, nil)
			Expect(err).NotTo(HaveOccurred())
			client := http.Client{}
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(200))
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodyBytes)).To(Equal("these are the bytes from proxied service."))
		})
		g.It("should proxy headers from remote", func() {
			proxyToThis := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-my-custom-header", "my custom header value")
				w.WriteHeader(200)
				w.Write([]byte("these are the bytes from proxied service."))
			}))
			logHandler := log.LogHandlerImpl{logger}

			server := httptest.NewServer(MakeProxiedHandler(proxyToThis.URL, &logHandler))
			req, err := http.NewRequest("GET", server.URL, nil)
			Expect(err).NotTo(HaveOccurred())
			client := http.Client{}
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			myCustomHeader := resp.Header.Get("X-my-custom-header")
			Expect(myCustomHeader).To(Equal("my custom header value"))
		})
		g.It("should proxy headers to remote", func() {
			recievedHeaderValue := ""
			proxyToThis := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				recievedHeaderValue = r.Header.Get("X-my-custom-header")
				w.WriteHeader(200)
				w.Write([]byte("these are the bytes from proxied service."))
			}))
			logHandler := log.LogHandlerImpl{logger}

			server := httptest.NewServer(MakeProxiedHandler(proxyToThis.URL, &logHandler))
			req, err := http.NewRequest("GET", server.URL, nil)
			req.Header.Add("X-my-custom-header", "my custom header value")
			Expect(err).NotTo(HaveOccurred())
			client := http.Client{}
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(recievedHeaderValue).To(Equal("my custom header value"))
		})
		g.It("should forward http body", func() {
			var bodybytes []byte
			proxyToThis := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				bodybytes, _ = ioutil.ReadAll(r.Body)
				w.WriteHeader(200)
				w.Write([]byte("these are the bytes from proxied service."))
			}))
			logHandler := log.LogHandlerImpl{logger}

			server := httptest.NewServer(MakeProxiedHandler(proxyToThis.URL, &logHandler))
			req, err := http.NewRequest("GET", server.URL, bytes.NewBuffer([]byte("This is my POST body")))
			Expect(err).NotTo(HaveOccurred())
			client := http.Client{}
			_, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bodybytes)).To(Equal("This is my POST body"))
		})
		g.It("should use given path against proxied resource", func() {
			var receivedPath string
			proxyToThis := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedPath = r.URL.Path
				w.WriteHeader(200)
				w.Write([]byte("these are the bytes from proxied service."))
			}))
			logHandler := log.LogHandlerImpl{logger}
			server := httptest.NewServer(MakeProxiedHandler(proxyToThis.URL, &logHandler))
			req, err := http.NewRequest("GET", server.URL+"/some/path", nil)
			Expect(err).NotTo(HaveOccurred())
			client := http.Client{}
			_, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(receivedPath).To(Equal("/some/path"))
		})
		g.It("should use given method against proxied resource", func() {
			var receivedMethod string
			proxyToThis := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedMethod = r.Method
				w.WriteHeader(200)
				w.Write([]byte("these are the bytes from proxied service."))
			}))
			logHandler := log.LogHandlerImpl{logger}
			server := httptest.NewServer(MakeProxiedHandler(proxyToThis.URL, &logHandler))
			for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
				client := http.Client{}
				req, err := http.NewRequest(method, server.URL, nil)
				_, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				Expect(receivedMethod).To(Equal(method))
			}
		})
	})
}
