package proxy

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/route"
)

type TCPProxy interface {
	Serve(conn net.Conn)
}

func NewTCPSNIProxy(cfg config.Proxy) TCPProxy {
	return &tcpSNIProxy{cfg: cfg}
}

type tcpSNIProxy struct {
	cfg config.Proxy
}

func (p *tcpSNIProxy) Serve(in net.Conn) {
	defer in.Close()

	if ShuttingDown() {
		return
	}

	// capture client hello
	data := make([]byte, 1024)
	n, err := in.Read(data)
	if err != nil {
		return
	}
	data = data[:n]

	serverName, ok := readServerName(data)
	if !ok {
		// println("handshake failed")
		fmt.Fprintln(in, "handshake failed")
		return
	}

	if serverName == "" {
		// println("server name missing")
		fmt.Fprintln(in, "server_name missing")
		return
	}
	// println(serverName)

	t := route.GetTable().LookupHost(serverName)
	if t == nil {
		log.Print("[WARN] No route for ", serverName)
		return
	}
	// println(serverName + " -> " + t.URL.Host)

	out, err := net.DialTimeout("tcp", t.URL.Host, p.cfg.DialTimeout)
	if err != nil {
		log.Println("cannot connect upstream")
		return
	}
	defer out.Close()
	// TODO(fs): set timeouts

	// copy client hello
	_, err = out.Write(data)
	if err != nil {
		log.Println("copy client hello failed")
		return
	}

	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}

	go cp(out, in)
	go cp(in, out)
	err = <-errc
	if err != nil && err != io.EOF {
		log.Println("error ", err)
	}
}
