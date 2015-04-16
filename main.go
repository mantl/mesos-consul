package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/CiscoCloud/mesos-consul/registry"
	"github.com/CiscoCloud/mesos-consul/mesos"

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

	var registryURI = flag.String("registry", "", "Registry URI")
	var refresh = flag.Duration("refresh", time.Minute, "Refresh duration")
	var zk = flag.String("zk", "", "Zookeeper address")

	flag.Parse()

	if *zk == "" {
		flag.Usage()
		os.Exit(1)
	}

	bridge := new(Bridge)

	bridge.registry = registry.GetRegistry(*registryURI)

	log.Print("Using zookeeper: ", *zk)
	bridge.leader = mesos.New(*zk, bridge.registry)

	ticker := time.NewTicker(*refresh)
        bridge.leader.Refresh()
	for _ = range ticker.C {
	        bridge.leader.Refresh()
	}
}
