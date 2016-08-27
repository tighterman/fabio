// Package metrics provides functions for collecting
// and managing metrics through different metrics libraries.
//
// Metrics library implementations must implement the
// Provider interface in the metricslib package.
//
// The current implementation supports only a single
// metrics provider.
package metrics

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/exit"
	"github.com/eBay/fabio/metrics/gometrics"
	"github.com/eBay/fabio/metrics/metricslib"
)

// Provider stores the metrics library provider.
var Provider metricslib.Provider = metricslib.NoopProvider{}

// Names returns the list of registered metrics acquired
// through the GetXXX() functions in alphabetical order.
func Names() []string { return Provider.Names() }

// Unregister removes the registered metric and stops
// reporting it to an external backend.
func Unregister(name string) { Provider.Unregister(name) }

// UnregisterAll removes all registered metrics and stops
// reporting  them to an external backend.
func UnregisterAll() { Provider.UnregisterAll() }

// GetTimer returns a timer metric for the given name.
// If the metric does not exist yet it should be created
// otherwise the existing metric should be returned.
func GetTimer(name string) metricslib.Timer { return Provider.GetTimer(name) }

// prefix stores the prefix for all metrics names.
var prefix string

// Init configures the metrics library provider and starts reporting.
func Init(cfg config.Metrics) error {
	prefix = cfg.Prefix
	if prefix == "default" {
		prefix = defaultPrefix()
	}

	var err error
	switch cfg.Target {
	case "stdout":
		log.Printf("[INFO] Sending metrics to stdout")
		Provider, err = gometrics.StdoutProvider(cfg.Interval)

	case "graphite":
		log.Printf("[INFO] Sending metrics to Graphite on %s as %q", cfg.GraphiteAddr, prefix)
		Provider, err = gometrics.GraphiteProvider(prefix, cfg.GraphiteAddr, cfg.Interval)

	case "statsd":
		log.Printf("[INFO] Sending metrics to StatsD on %s as %q", cfg.StatsDAddr, prefix)
		Provider, err = gometrics.StatsDProvider(prefix, cfg.StatsDAddr, cfg.Interval)

	case "":
		log.Printf("[INFO] Metrics disabled")

	default:
		exit.Fatal("[FATAL] Invalid metrics target ", cfg.Target)
	}
	return err
}

// TargetName returns the metrics name from the given parameters.
func TargetName(service, host, path string, targetURL *url.URL) string {
	return strings.Join([]string{
		clean(service),
		clean(host),
		clean(path),
		clean(targetURL.Host),
	}, ".")
}

// clean creates safe names for graphite reporting by replacing
// some characters with underscores.
// TODO(fs): This may need updating for other metrics backends.
func clean(s string) string {
	if s == "" {
		return "_"
	}
	s = strings.Replace(s, ".", "_", -1)
	s = strings.Replace(s, ":", "_", -1)
	return strings.ToLower(s)
}

// stubbed out for testing
var hostname = os.Hostname

// defaultPrefix determines the default metrics prefix from
// the current hostname and the name of the executable.
func defaultPrefix() string {
	host, err := hostname()
	if err != nil {
		exit.Fatal("[FATAL] ", err)
	}
	exe := filepath.Base(os.Args[0])
	return clean(host) + "." + clean(exe)
}
