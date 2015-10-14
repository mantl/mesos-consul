package config

import (
	"time"
)

type Config struct {
	Refresh       time.Duration
	Zk            string
	LogLevel      string
}

func DefaultConfig() *Config {
	return &Config{
		Refresh: time.Minute,
		Zk:            "zk://127.0.0.1:2181/mesos",
	}
}
