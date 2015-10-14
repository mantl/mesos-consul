package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/CiscoCloud/mesos-consul/config"
	"github.com/CiscoCloud/mesos-consul/consul"
	"github.com/CiscoCloud/mesos-consul/mesos"

	flag "github.com/ogier/pflag"
	log "github.com/sirupsen/logrus"
)

const Name = "mesos-consul"
const Version = "0.3"

func main() {
	c, err := parseFlags(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Using zookeeper: ", c.Zk)
	leader := mesos.New(c)

	ticker := time.NewTicker(c.Refresh)
	leader.Refresh()
	for _ = range ticker.C {
		leader.Refresh()
	}
}

func parseFlags(args []string) (*config.Config, error) {
	var doHelp bool
	var c = config.DefaultConfig()

	flags := flag.NewFlagSet("mesos-consul", flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Println(Help())
	}

	flags.BoolVar(&doHelp, "help", false, "")
	flags.StringVar(&c.LogLevel, "log-level", "WARN", "")
	flags.DurationVar(&c.Refresh, "refresh", time.Minute, "")
	flags.StringVar(&c.Zk, "zk", "zk://127.0.0.1:2181/mesos", "")
	flags.StringVar(&c.MesosIpOrder, "mesos-ip-order", "netinfo,mesos,host", "")

	consul.AddCmdFlags(flags)

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	args = flags.Args()
	if len(args) > 0 {
		return nil, fmt.Errorf("extra argument(s): %q", args)
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

  --log-level=<log_level>	Set the Logging level to one of [ "DEBUG", "INFO", "WARN", "ERROR" ]
				(default "WARN")
  --refresh=<time>		Set the Mesos refresh rate
				(default 1m)
  --zk=<address>		Zookeeper path to Mesos
				(default zk://127.0.0.1:2181/mesos)
  --mesos-ip-order=		Comma separated list to control the order in
				which mesos-consul searches for the task IP
				address. Valid options are 'netinfo', 'mesos',
				'docker' and 'host'
				(default netinfo,mesos,host)
` + consul.Help()

	return strings.TrimSpace(helpText)
}
