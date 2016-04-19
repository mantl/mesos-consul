package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/CiscoCloud/mesos-consul/consul"
	"github.com/CiscoCloud/mesos-consul/mesos"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const Name = "mesos-consul"
const Version = "0.4.0"

func main() {
	root := &cobra.Command{
		Use: "mesos-consul",
		Short: "Mesos to Consul bridge for service discovery",
		Long: "Mesos to Consul bridge for service discovery",
		PreRun: func (cmd *cobra.Command, args []string) {
			lev := viper.GetString("log-level")
			l, err := log.ParseLevel(strings.ToLower(lev))
			if err != nil {
				log.SetLevel(log.WarnLevel)
				log.Warnf("Invalid log level '%v'. Setting to WARN", )
			} else {
				log.SetLevel(l)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			DoIt()
		},
	}

	root.Flags().Bool("version", false, "Print mesos-consul version")
	root.Flags().String("log-level", "WARN", `Set the log level to one of ["DEBUG","INFO","WARN", "ERROR"]`)
	root.Flags().Duration("refresh", time.Minute, "Set the Mesos refresh rate")
	root.Flags().String("zk", "zk://127.0.0.1:2181/mesos", "Zookeeper path to Mesos")
	root.Flags().String("group-separator", "", "Choose the group separator. Will replace _ in task names")
	root.Flags().String("mesos-ip-order", "netinfo,mesos,host", "Comma separated list to control the order in which mesos-consul searches for the task IP address. Valid options are 'netinfo', 'mesos', 'docker' and 'host'")
	root.Flags().Bool("healthcheck", false, "Enables the http endpoint for health checks")
	root.Flags().String("healthcheck-ip", "127.0.0.1", "Health check interface IP")
	root.Flags().String("healthcheck-port", "24476", "Health check interface Port")

	root.Flags().StringSlice("whitelist", nil, "Only register services matching the provided regex. Can be specified multiple times")
	root.Flags().StringSlice("blacklist", nil, "Do not register services matching the provided regex. Can be specified multiple times")
	root.Flags().StringSlice("fw-whitelist", nil, "Only register services from frameworks matching the provided regex. Can be specified multiple times")
	root.Flags().StringSlice("fw-blacklist", nil, "Do not register services from frameworks matching the provided regex. Can be specified multiple times")
	root.Flags().StringSlice("task-tag", nil, "Tag tasks whose name contains 'pattern' substring (case-insensitive) with given tag. Can be specified multiple times")

	root.Flags().String("service-name", "mesos", "Service name of the Mesos hosts")
	root.Flags().String("service-tags", "", "Comma delimited list of tags to add to the mesos hosts. Hosts are registered as (leader|master|follower).<tag>.mesos.service.consul")

	consul.InitFlags(root)

	viper.BindPFlags(root.Flags())

	root.Execute()
}

func DoIt() {
	if viper.GetBool("healthcheck") {
		go StartHealthcheckService(viper.GetString("healthcheck-ip"), viper.GetString("healthcheck-port"))
	}

	log.Info("Using zookeeper: ", viper.GetString("zk"))
	leader := mesos.New()

	ticker := time.NewTicker(viper.GetDuration("refresh"))
	leader.Refresh()
	for _ = range ticker.C {
		leader.Refresh()
	}
}

func StartHealthcheckService(ip string, port string) {
	http.HandleFunc("/health", HealthHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", ip, port), nil))
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}
