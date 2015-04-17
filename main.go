package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/CiscoCloud/mesos-consul/registry"
	"github.com/CiscoCloud/mesos-consul/mesos"
	"github.com/CiscoCloud/mesos-consul/config"

	flag "github.com/ogier/pflag"
)

type Bridge struct {
	registry	registry.RegistryAdapter		
	leader		*mesos.Mesos
}

func Usage() {
	fmt.Println("Usage: mesos-consul --zk=\"\" --registry=\"\" [options]\n")
	fmt.Println("Options:")
	fmt.Println("	--refresh=		Refresh time (default 1m)")
	fmt.Println()
}

func main() {

	c, err := parseFlags(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	bridge := new(Bridge)

	bridge.registry = registry.GetRegistry(c)

	log.Print("Using zookeeper: ", c.Zk)
	bridge.leader = mesos.New(c, bridge.registry)

	ticker := time.NewTicker(c.Refresh)
        bridge.leader.Refresh()
	for _ = range ticker.C {
	        bridge.leader.Refresh()
	}
}

func parseFlags(args []string) (*config.Config, error) {
	var c = config.DefaultConfig()

	flags := flag.NewFlagSet("mesos-consul", flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Print(usage)
	}

	flags.DurationVar(&c.Refresh,		"refresh", time.Minute, "")
	flags.StringVar(&c.Registry,		"registry", "", "")
	flags.Var((*config.AuthVar)(c.RegistryAuth),	"registry-auth", "")
	flags.BoolVar(&c.RegistrySSL.Enabled,	"registry-ssl", c.RegistrySSL.Enabled, "")
	flags.BoolVar(&c.RegistrySSL.Verify,	"registry-ssl-verify", c.RegistrySSL.Verify, "")
	flags.StringVar(&c.RegistrySSL.Cert,	"registry-ssl-cert", c.RegistrySSL.Cert, "")
	flags.StringVar(&c.RegistrySSL.CaCert,	"registry-ssl-cacert", c.RegistrySSL.CaCert, "")
	flags.StringVar(&c.RegistryToken,		"registry-token", c.RegistryToken, "")
	flags.StringVar(&c.Zk,			"zk", "", "")

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	args = flags.Args()
	if len(args) > 0 {
		return nil, fmt.Errorf("extra argument(s): %q", args)
	}

	if c.Registry == "" {
		return nil, fmt.Errorf("registry address not provided")
	}

	if c.Zk == "" {
		return nil, fmt.Errorf("Zookeeper address not provided")
	}

	return c, nil
}

const usage = `
Usage: mesos-consul [options]

Options:

  --refresh=<time>			Set the Mesos refresh rate (default 1m)
  --registry=<address>			Set the registry address
  --registry-auth=<user[:pass]>		Set the basic authentication username
					(and password)
  --registry-ssl			Use SSL when connecting to the registry
  --registry-ssl-verify			Verify certificates when connecting
					via SSL
  --registry-ssl-cert			SSL certificates to send to registry
  --registry-ssl-cacert			Validate server certificate against
					this CA
					certificate file list
  --registry-token=<token>		Set registry ACL token
  --zk=<address>			Zookeeper path to Mesos
`
