// Package gometrics provides an implementation of the
// metricslib.Provider interface using the github.com/rcrowley/go-metrics
// library.
package gometrics

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"time"

	"github.com/eBay/fabio/metrics/metricslib"

	graphite "github.com/cyberdelia/go-metrics-graphite"
	statsd "github.com/pubnub/go-metrics-statsd"
	gm "github.com/rcrowley/go-metrics"
)

// StdoutProvider returns a provider that reports to stdout.
func StdoutProvider(interval time.Duration) (metricslib.Provider, error) {
	registry := gm.NewRegistry()
	logger := log.New(os.Stderr, "localhost: ", log.Lmicroseconds)
	go gm.Log(gm.DefaultRegistry, interval, logger)
	go gm.Log(registry, interval, logger)
	return &gmProvider{registry}, nil
}

// GraphiteProvider returns a provider that reports to a Graphite server.
func GraphiteProvider(prefix, addr string, interval time.Duration) (metricslib.Provider, error) {
	if addr == "" {
		return nil, errors.New("metrics: graphite addr missing")
	}

	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("metrics: cannot connect to Graphite: %s", err)
	}

	registry := gm.NewRegistry()
	go graphite.Graphite(gm.DefaultRegistry, interval, prefix, a)
	go graphite.Graphite(registry, interval, prefix, a)
	return &gmProvider{registry}, nil
}

// StatsDProvider returns a provider that reports to a StatsD server.
func StatsDProvider(prefix, addr string, interval time.Duration) (metricslib.Provider, error) {
	if addr == "" {
		return nil, errors.New("metrics: statsd addr missing")
	}

	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("metrics: cannot connect to StatsD: %s", err)
	}

	registry := gm.NewRegistry()
	go statsd.StatsD(gm.DefaultRegistry, interval, prefix, a)
	go statsd.StatsD(registry, interval, prefix, a)
	return &gmProvider{registry}, nil
}

// gmProvider implements the metricslib.Provider interface
// using the github.com/rcrowley/go-metrics library.
type gmProvider struct {
	// registry keeps track of registered metrics values
	// to keep them separate from the default metrics.
	registry gm.Registry
}

func (p *gmProvider) Names() (names []string) {
	p.registry.Each(func(name string, _ interface{}) {
		names = append(names, name)
	})
	sort.Strings(names)
	return names
}

func (p *gmProvider) Unregister(name string) {
	p.registry.Unregister(name)
}

func (p *gmProvider) UnregisterAll() {
	p.registry.UnregisterAll()
}

func (p *gmProvider) GetTimer(name string) metricslib.Timer {
	return gm.GetOrRegisterTimer(name, p.registry)
}
