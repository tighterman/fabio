package logger

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	t1 := time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := t1.Add(123456 * time.Microsecond)

	req := &http.Request{
		RequestURI: "/?q=x",
		Header: http.Header{
			"User-Agent":      {"Mozilla Firefox"},
			"X-Forwarded-For": {"3.3.3.3"},
		},
		RemoteAddr: "2.2.2.2:666",
		Host:       "server host",
		URL: &url.URL{
			Host: "proxy host",
		},
		Method: "GET",
		Proto:  "HTTP/1.1",
	}

	tests := []struct {
		format string
		out    string
	}{
		{"remote_addr", "2.2.2.2\n"},
		{"time", "2016-01-01T00:00:00Z\n"},
		//{"request", "\"GET /?q=x HTTP/1.1\"\n"},
		{"body_bytes_sent", "0\n"},
		{"http_referer", "\n"},
		{"http_user_agent", "Mozilla Firefox\n"},
		{"http_x_forwarded_for", "3.3.3.3\n"},
		{"server_name", "server host\n"},   // TODO(fs): is this correct?
		{"proxy_endpoint", "proxy host\n"}, // TODO(fs): is this correct?
		{"response_time", "0.123456\n"},    // TODO(fs): is this correct?
		//{"request_args", "?q=x\n"},
	}

	for i, tt := range tests {
		b := new(bytes.Buffer)

		l, err := New(b, tt.format)
		if err != nil {
			t.Fatalf("%d: got %v want nil", i, err)
		}

		l.Log(t1, t2, req)

		if got, want := string(b.Bytes()), tt.out; got != want {
			t.Errorf("%d: got %q want %q", i, got, want)
		}
	}
}

func TestAtoi(t *testing.T) {
	tests := []struct {
		i   int64
		pad int
		s   string
	}{
		{i: 0, pad: 0, s: "0"},
		{i: 1, pad: 0, s: "1"},
		{i: -1, pad: 0, s: "-1"},
		{i: 12345, pad: 0, s: "12345"},
		{i: -12345, pad: 0, s: "-12345"},
		{i: 9223372036854775807, pad: 0, s: "9223372036854775807"},
		{i: -9223372036854775807, pad: 0, s: "-9223372036854775807"},

		{i: 0, pad: 5, s: "00000"},
		{i: 1, pad: 5, s: "00001"},
		{i: -1, pad: 5, s: "-00001"},
		{i: 12345, pad: 5, s: "12345"},
		{i: -12345, pad: 5, s: "-12345"},
		{i: 9223372036854775807, pad: 5, s: "9223372036854775807"},
		{i: -9223372036854775807, pad: 5, s: "-9223372036854775807"},
	}

	for i, tt := range tests {
		var b bytes.Buffer
		atoi(&b, tt.i, tt.pad)
		if got, want := string(b.Bytes()), tt.s; got != want {
			t.Errorf("%d: got %q want %q", i, got, want)
		}
	}
}

func BenchmarkLog(b *testing.B) {
	t1 := time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := t1.Add(100 * time.Millisecond)
	req := &http.Request{
		RequestURI: "/",
		Header:     http.Header{"X-Forwarded-For": {"3.3.3.3"}},
		RemoteAddr: "2.2.2.2:666",
		URL:        &url.URL{},
		Method:     "GET",
		Proto:      "HTTP/1.1",
	}

	var keys []string
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	format := strings.Join(keys, " ")

	l, err := New(ioutil.Discard, format)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		l.Log(t1, t2, req)
	}
}
