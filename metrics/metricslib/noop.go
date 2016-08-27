package metricslib

import "time"

// NoopProvider is a stub implementation of the Provider interface.
type NoopProvider struct{}

func (p NoopProvider) Names() []string { return nil }

func (p NoopProvider) Unregister(name string) {}

func (p NoopProvider) UnregisterAll() {}

func (p NoopProvider) GetTimer(name string) Timer { return noopTimer }

var noopTimer = NoopTimer{}

// NoopTimer is a stub implementation of the Timer interface.
type NoopTimer struct{}

func (t NoopTimer) UpdateSince(start time.Time) {}

func (t NoopTimer) Rate1() float64 { return 0 }

func (t NoopTimer) Percentile(nth float64) float64 { return 0 }
