package consul

import (
	"fmt"
	"strings"

	flag "github.com/ogier/pflag"
)

type consulConfig struct {
	enabled		bool
        auth            auth
        port            string
        sslEnabled      bool
        sslVerify       bool
        sslCert         string
        sslCaCert       string
        token           string
}

// XXX -Rename to config after removing config import
var config consulConfig

func AddCmdFlags(f *flag.FlagSet) {
	f.BoolVar(&config.enabled,	"consul", false, "")
        f.StringVar(&config.port,      "consul-port", "8500", "")
        f.Var((*authVar)(&config.auth), "consul-auth", "")
        f.BoolVar(&config.sslEnabled, "consul-ssl", false, "")
        f.BoolVar(&config.sslVerify, "consul-ssl-verify", true, "")
        f.StringVar(&config.sslCert, "consul-ssl-cert", "", "")
        f.StringVar(&config.sslCaCert, "consul-ssl-cacert", "", "")
        f.StringVar(&config.token, "consul-token", "", "")
}

func Help() string {
	helpText := `
Consul Options:

  --consul			Use Consul backend
  --consul-port			Consul agent API port
				(default: 8500)
  --consul-auth			The basic authentication username (and optional password),
				separated by a colon.
				(default: not set)
  --consul-ssl			Use HTTPS when talking to Consul
				(default: false)
  --consul-ssl-verify		Verify certificates when connecting via SSL
				(default: true)
  --consul-ssl-cert		Path to an SSL client certificate to use to authenticate
				to the Consul server
				(default: not set)
  --consul-ssl-cacert		Path to a CA certificate file, containing one or more CA
				certificates to use to validate the certificate sent
				by the Consul server to us
				(default: not set)
  --consul-token		The Consul ACL token
				(default: not set)

`

	return helpText
}

func IsEnabled() bool {
	return config.enabled
}

type auth struct {
	Enabled		bool
	Username	string
	Password	string
}

// AuthVar implements the Flag.Value interface and allows the user to specify
// authentication in the username[:password] form.
type authVar auth

func (a *authVar) Set(value string) error {
	a.Enabled = true

	if (strings.Contains(value, ":")) {
		split := strings.SplitN(value, ":", 2)
		a.Username = split[0]
		a.Password = split[1]
	} else {
		a.Username = value
	}

	return nil
}

func (a *authVar) String() string {
	if a.Password == "" {
		return a.Username
	}

	return fmt.Sprintf("%s:%s", a.Username, a.Password)
}
