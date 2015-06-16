package proxy

import (
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
	})
}