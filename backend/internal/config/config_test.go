package config

import (
	"os"
	"testing"
)

func TestLoad_UsesDefaultsWhenUnset(t *testing.T) {
	t.Setenv("PORT", "")
	os.Unsetenv("PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != defaultPort {
		t.Errorf("Port = %d, want %d", cfg.Port, defaultPort)
	}
	if cfg.ReadTimeout != readTimeout || cfg.WriteTimeout != writeTimeout || cfg.IdleTimeout != idleTimeout {
		t.Errorf("timeouts = %v/%v/%v, want the defaults", cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout)
	}
}

func TestLoad_TreatsAnEmptyPortAsUnset(t *testing.T) {
	t.Setenv("PORT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != defaultPort {
		t.Errorf("Port = %d, want %d", cfg.Port, defaultPort)
	}
}

func TestLoad_OverridesThePort(t *testing.T) {
	t.Setenv("PORT", "8090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8090 {
		t.Errorf("Port = %d, want 8090", cfg.Port)
	}
}

func TestLoad_RejectsInvalidPorts(t *testing.T) {
	tests := []struct {
		name string
		port string
	}{
		{name: "not a number", port: "http"},
		{name: "decimal", port: "9080.5"},
		{name: "zero", port: "0"},
		{name: "negative", port: "-1"},
		{name: "above the maximum", port: "65536"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PORT", tt.port)

			if _, err := Load(); err == nil {
				t.Errorf("Load() with PORT=%q returned no error", tt.port)
			}
		})
	}
}

func TestAddress(t *testing.T) {
	if got := (Config{Port: 9080}).Address(); got != ":9080" {
		t.Errorf("Address() = %q, want %q", got, ":9080")
	}
}
