package route

import (
	"reflect"
	"testing"

	"github.com/eBay/fabio/metrics"
	"github.com/eBay/fabio/metrics/metricslib"
)

func TestSyncRegistry(t *testing.T) {
	oldProvider := metrics.Provider
	metrics.Provider = newStubProvider()
	defer func() { metrics.Provider = oldProvider }()

	tbl := make(Table)
	tbl.AddRoute("svc-a", "/aaa", "http://localhost:1234", 1, nil)
	tbl.AddRoute("svc-b", "/bbb", "http://localhost:5678", 1, nil)
	if got, want := metrics.Names(), []string{"svc-a._./aaa.localhost_1234", "svc-b._./bbb.localhost_5678"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	tbl.DelRoute("svc-b", "/bbb", "http://localhost:5678")
	syncRegistry(tbl)
	if got, want := metrics.Names(), []string{"svc-a._./aaa.localhost_1234"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func newStubProvider() metricslib.Provider {
	return &stubProvider{names: make(map[string]bool)}
}

type stubProvider struct {
	names map[string]bool
}

func (p *stubProvider) Names() []string {
	n := []string{}
	for k := range p.names {
		n = append(n, k)
	}
	return n
}

func (p *stubProvider) Unregister(name string) {
	delete(p.names, name)
}

func (p *stubProvider) UnregisterAll() {
	p.names = map[string]bool{}
}

func (p *stubProvider) GetTimer(name string) metricslib.Timer {
	p.names[name] = true
	return metricslib.NoopTimer{}
}
