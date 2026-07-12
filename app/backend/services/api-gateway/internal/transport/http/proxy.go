package http

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	proxy *httputil.ReverseProxy
}

func NewProxyHandler(target *url.URL) *ProxyHandler {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":{"code":"bad_gateway","message":"upstream service unavailable"}}`))
	}
	return &ProxyHandler{proxy: proxy}
}

func (h *ProxyHandler) Serve(c *gin.Context) {
	h.proxy.ServeHTTP(c.Writer, c.Request)
}
