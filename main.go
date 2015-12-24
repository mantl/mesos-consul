package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"net/http"

	"github.com/CiscoCloud/mesos-consul/config"
	"github.com/CiscoCloud/mesos-consul/consul"
	"github.com/CiscoCloud/mesos-consul/mesos"

	flag "github.com/ogier/pflag"
	log "github.com/sirupsen/logrus"
)

const Name = "mesos-consul"
const Version = "0.3.1"

func main() {
	c, err := parseFlags(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if c.Healthcheck {
		go StartHealthcheckService(c)
	}

	log.Info("Using zookeeper: ", c.Zk)
	leader := mesos.New(c)

	ticker := time.NewTicker(c.Refresh)
	leader.Refresh()
	for _ = range ticker.C {
		leader.Refresh()
	}
}

func StartHealthcheckService(c *config.Config) {
	http.HandleFunc("/health", HealthHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", c.HealthcheckIp, c.HealthcheckPort), nil))
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintln(w, "OK")
}

func parseFlags(args []string) (*config.Config, error) {
	var doHelp bool
	var doVersion bool
	var c = config.DefaultConfig()

	flags := flag.NewFlagSet("mesos-consul", flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Println(Help())
	}

	flags.BoolVar(&doHelp, "help", false, "")
	flags.BoolVar(&doVersion, "version", false, "")
	flags.StringVar(&c.LogLevel, "log-level", "WARN", "")
	flags.DurationVar(&c.Refresh, "refresh", time.Minute, "")
	flags.StringVar(&c.Zk, "zk", "zk://127.0.0.1:2181/mesos", "")
	flags.StringVar(&c.MesosIpOrder, "mesos-ip-order", "netinfo,mesos,host", "")
	flags.BoolVar(&c.Healthcheck, "healthcheck", false, "")
	flags.StringVar(&c.HealthcheckIp, "healthcheck-ip", "127.0.0.1", "")
	flags.StringVar(&c.HealthcheckPort, "healthcheck-port", "24476", "")
	flags.Var((funcVar)(func(s string) error {
		c.WhiteList = append(c.WhiteList, s)
		return nil
	}), "whitelist", "")
	flags.StringVar(&c.ServiceName, "service-name", "mesos", "")
	flags.StringVar(&c.ServiceTags, "service-tags", "", "")

	consul.AddCmdFlags(flags)

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	args = flags.Args()
	if len(args) > 0 {
		return nil, fmt.Errorf("extra argument(s): %q", args)
	}

	if doVersion {
		fmt.Printf("%s v%s\n", Name, Version)
		os.Exit(0)
	}
	if doHelp {
		flags.Usage()
		os.Exit(0)
	}

	l, err := log.ParseLevel(strings.ToLower(c.LogLevel))
	if err != nil {
		log.SetLevel(log.WarnLevel)
		log.Warnf("Invalid log level '%v'. Setting to WARN", c.LogLevel)
	} else {
		log.SetLevel(l)
	}

	return c, nil
}

func Help() string {
	helpText := `
Usage: mesos-consul [options]

Options:

  --version 			Print mesos-consul version
  --log-level=<log_level>	Set the Logging level to one of [ "DEBUG", "INFO", "WARN", "ERROR" ]
				(default "WARN")
  --refresh=<time>		Set the Mesos refresh rate (default 1m)
  --zk=<address>		Zookeeper path to Mesos (default zk://127.0.0.1:2181/mesos)
  --healthcheck 		Enables a http endpoint for health checks. When this
				flag is enabled, serves a service health status on 127.0.0.1:24476 (default not enabled)
  --healthcheck-ip=<ip> 	Health check interface ip (default 127.0.0.1)
  --healthcheck-port=<port>	Health check service port (default 24476)
  --mesos-ip-order		Comma separated list to control the order in
				which github.com/CiscoCloud/mesos-consul searches for the task IP
				address. Valid options are 'netinfo', 'mesos', 'docker' and 'host'
				(default netinfo,mesos,host)
` + consul.Help()

	return strings.TrimSpace(helpText)
}

type funcVar func(s string) error

func (f funcVar) Set(s string) error { return f(s) }
func (f funcVar) String() string { return "" }
func (f funcVar) IsBoolFlag() bool {return false }
