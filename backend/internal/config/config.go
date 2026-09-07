// Package config reads the server's runtime settings from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// 8080 is commonly already claimed on a developer machine, so the service defaults to 9080 and
// stays overridable rather than failing to bind for a reason that is hard to see.
const (
	defaultPort  = 9080
	minPort      = 1
	maxPort      = 65535
	readTimeout  = 5 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 60 * time.Second
)

type Config struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Address renders the listen address for net/http.
func (c Config) Address() string {
	return ":" + strconv.Itoa(c.Port)
}

// Load builds the configuration from the environment, applying defaults where nothing is set.
func Load() (Config, error) {
	port, err := intFromEnv("PORT", defaultPort)
	if err != nil {
		return Config{}, err
	}
	if port < minPort || port > maxPort {
		return Config{}, fmt.Errorf("PORT must be between %d and %d, got %d", minPort, maxPort, port)
	}

	return Config{
		Port:         port,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}, nil
}

func intFromEnv(key string, fallback int) (int, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a whole number, got %q", key, raw)
	}
	return value, nil
}
