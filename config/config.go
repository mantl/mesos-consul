package config

import (
	"time"
)

type Config struct {
	Refresh          time.Duration
	Zk               string
	LogLevel         string
	MesosIpOrder     string
	Healthcheck      bool
	HealthcheckIp    string
	HealthcheckPort  string
}

func DefaultConfig() *Config {
	return &Config{
		Refresh:        time.Minute,
		Zk:             "zk://127.0.0.1:2181/mesos",
		MesosIpOrder:   "netinfo,mesos,host",
		Healthcheck:    false,
		HealthcheckIp:  "127.0.0.1",
		HealthcheckPort: "24476",
	}
}
