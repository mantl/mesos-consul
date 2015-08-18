package config

import (
	"time"
)

type Auth struct {
	Enabled  bool
	Username string
	Password string
}

type SSL struct {
	Enabled bool
	Verify  bool
	Cert    string
	CaCert  string
}

type Config struct {
	Refresh       time.Duration
	RegistryAuth  *Auth
	RegistryPort  string
	RegistrySSL   *SSL
	RegistryToken string
	Zk            string
	LogLevel      string
}

func DefaultConfig() *Config {
	return &Config{
		Refresh: time.Minute,
		RegistryAuth: &Auth{
			Enabled: false,
		},
		RegistrySSL: &SSL{
			Enabled: false,
			Verify:  true,
		},
		RegistryToken: "",
		Zk:            "zk://127.0.0.1:2181/mesos",
	}
}
