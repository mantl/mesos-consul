package consul

import (
	"strings"

	"github.com/spf13/cobra"
)

type consulConfig struct {
	enabled                bool
	auth                   auth
	port                   string
	sslEnabled             bool
	sslVerify              bool
	sslCert                string
	sslCaCert              string
	token                  string
	heartbeatsBeforeRemove int
}

var config consulConfig

func InitFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("consul", false, "Use Consul backend")
	cmd.Flags().String("consul-port", "8500", "Consul agent API port")
	cmd.Flags().String("consul-auth", "", "The basic authentication username (and optional password) separated by a colon")
	cmd.Flags().Bool("consul-ssl", false, "Use HTTPS when talking to Consul")
	cmd.Flags().Bool("consul-ssl-verify", true, "Verify certificates when connecting via SSL")
	cmd.Flags().String("consul-ssl-cert", "", "Path to an SSL client certificate to use to authenticate to the Consul server")
	cmd.Flags().String("consul-ssl-cacert", "", "Path to a CA certificate file, containing one or more CA certificates to use to validate the certificate sent by the Consul server to us")
	cmd.Flags().String("consul-token", "", "The Consul ACL token")
	cmd.Flags().Int("heartbeats-before-remove", 1, "Number of times that registration needs to fail before removing task from Consul")
}

type auth struct {
	Enabled  bool
	Username string
	Password string
}

func toAuth(s string) auth {
	var a auth

	a.Enabled = true

	if strings.Contains(s, ":") {
		split := strings.SplitN(s, ":", 2)
		a.Username = split[0]
		a.Password = split[1]
	} else {
		a.Username = s
	}

	return a
}
