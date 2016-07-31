// Package logger implements a configurable access logger.
package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type field func(b *bytes.Buffer, t1, t2 time.Time, r *http.Request)

var fields = map[string]field{
	"remote_addr": func(b *bytes.Buffer, t1, t2 time.Time, r *http.Request) {
		b.WriteString(r.RemoteAddr[:strings.Index(r.RemoteAddr, ":")])
	},
	"time": func(b *bytes.Buffer, t1, t2 time.Time, r *http.Request) {
		atoi(b, int64(t2.Year()), 4)
		b.WriteRune('-')
		atoi(b, int64(t2.Month()), 2)
		b.WriteRune('-')
		atoi(b, int64(t2.Day()), 2)
		b.WriteRune('T')
		atoi(b, int64(t2.Hour()), 2)
		b.WriteRune(':')
		atoi(b, int64(t2.Minute()), 2)
		b.WriteRune(':')
		atoi(b, int64(t2.Second()), 2)
		b.WriteRune('Z')
	},
	"request": func(b *bytes.Buffer, t1, t2 time.Time, r *http.Request) {
		b.WriteRune('"')
		b.WriteString(r.Method)
		b.WriteRune(' ')
		b.WriteString(r.RequestURI)
		b.WriteRune(' ')
		b.WriteString(r.Proto)
		b.WriteRune('"')
	},
	"body_bytes_sent": func(b *bytes.Buffer, t1, t2 time.Time, r *http.Request) {
		atoi(b, r.ContentLength, 0)
	},
	"http_referer": func(b *bytes.Buffer, t1, t2 time.Time, r *http.Request) {
		b.WriteString(r.Referer())
	},
	"http_user_agent": func(b *bytes.Buffer, t1, t2 time.Time, r *http.Request) {
		b.WriteString(r.UserAgent())
	},
	"http_x_forwarded_for": func(b *bytes.Buffer, t1, t2 time.Time, r *http.Request) {
		b.WriteString(r.Header.Get("X-Forwarded-For"))
	},
	"server_name": func(b *bytes.Buffer, t1, t2 time.Time, r *http.Request) {
		// TODO(fs): is this correct?
		b.WriteString(r.Host)
	},
	"proxy_endpoint": func(b *bytes.Buffer, t1, end time.Time, r *http.Request) {
		// TODO(fs): is this correct?
		b.WriteString(r.URL.Host)
	},
	"response_time": func(b *bytes.Buffer, t1, t2 time.Time, r *http.Request) {
		// TODO(fs): is this correct?
		d := t2.Sub(t1).Nanoseconds()
		s, µs := d/int64(time.Second), d%int64(time.Second)/int64(time.Microsecond)
		atoi(b, s, 0)
		b.WriteRune('.')
		atoi(b, µs, 6)
	},
	"request_args": func(b *bytes.Buffer, t1, t2 time.Time, r *http.Request) {
		b.WriteString(r.URL.RawQuery)
	},
}

type Logger struct {
	p []field

	// w is the log destination
	w io.Writer

	// mu guards w
	mu sync.Mutex
}

func New(w io.Writer, format string) (*Logger, error) {
	p, err := parse(format)
	if err != nil {
		return nil, err
	}
	if len(p) == 0 {
		return nil, fmt.Errorf("log: invalid format %q", format)
	}
	return &Logger{w: w, p: p}, nil
}

func parse(format string) (pattern []field, err error) {
	for _, f := range strings.Fields(format) {
		p := fields[f]
		if p == nil {
			return nil, fmt.Errorf("log: invalid field %q", f)
		}
		pattern = append(pattern, p)
	}
	return
}

const BufSize = 1024

var pool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, BufSize))
	},
}

func (l *Logger) Log(t1, t2 time.Time, r *http.Request) {
	b := pool.Get().(*bytes.Buffer)
	b.Reset()

	for _, p := range l.p {
		p(b, t1, t2, r)
		b.WriteRune(' ')
	}
	b.Truncate(b.Len() - 1) // drop last space
	b.WriteRune('\n')

	l.mu.Lock()
	l.w.Write(b.Bytes())
	l.mu.Unlock()
	pool.Put(b)
}

// atoi is a replacement for strconv.Atoi/strconv.FormatInt
// which does not alloc.
func atoi(b *bytes.Buffer, i int64, pad int) {
	var flag bool
	if i < 0 {
		flag = true
		i = -i
	}

	// format number
	// 2^63-1 == 9223372036854775807
	var d [128]byte
	n, p := len(d), len(d)-1
	for i >= 0 {
		d[p] = byte('0') + byte(i%10)
		i /= 10
		p--
		if i == 0 {
			break
		}
	}

	// padding
	for n-p-1 < pad {
		d[p] = byte('0')
		p--
	}

	if flag {
		d[p] = '-'
		p--
	}
	b.Write(d[p+1:])
}
