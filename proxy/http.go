package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	gometrics "github.com/rcrowley/go-metrics"
)

func newHTTPProxy(t *url.URL, tr http.RoundTripper) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(t)
	rp.Transport = &meteredRoundTripper{tr}
	return rp
}

type meteredRoundTripper struct {
	tr http.RoundTripper
}

func (m *meteredRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := m.tr.RoundTrip(r)
	gometrics.GetOrRegisterTimer(name(resp.StatusCode), gometrics.DefaultRegistry).UpdateSince(start)
	return resp, err
}

func name(code int) string {
	b := []byte("http.status.")
	b = strconv.AppendInt(b, int64(code), 10)
	return string(b)
}
