package mesos

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/CiscoCloud/mesos-consul/config"
	"github.com/CiscoCloud/mesos-consul/consul"

	consulapi "github.com/hashicorp/consul/api"
)

type CacheEntry struct {
	service      *consul.ServiceRegistration
	isRegistered bool
}

type Mesos struct {
	Consul       *consul.Consul
	Masters      *[]MesosHost
	Lock         sync.Mutex
	ServiceCache map[string]*CacheEntry
}

var ipLabels = map[string]string{
	"docker": "Docker.NetworkSettings.IPAddress",
	"mesos":  "MesosContainerizer.NetworkSettings.IPAddress",
}

func New(c *config.Config, consul *consul.Consul) *Mesos {
	m := new(Mesos)

	if c.Zk == "" {
		return nil
	}

	m.Consul = consul

	m.zkDetector(c.Zk)

	return m
}

func (m *Mesos) Refresh() error {
	sj, err := m.loadState()
	if err != nil {
		log.Print("[ERROR] No master")
		return err
	}

	if sj.Leader == "" {
		return errors.New("Empty master")
	}

	if m.ServiceCache == nil {
		log.Print("[INFO] Creating ServiceCache")
		m.ServiceCache = make(map[string]*CacheEntry)
		m.LoadCache()
	}

	m.parseState(sj)

	return nil
}

func (m *Mesos) loadState() (StateJSON, error) {
	var err error
	var sj StateJSON

	defer func() {
		if rec := recover(); rec != nil {
			err = errors.New("can't connect to Mesos")
		}
	}()

	ip, port := m.getLeader()
	if ip == "" {
		return sj, errors.New("No master in zookeeper")
	}

	log.Printf("[INFO] Zookeeper leader: %s:%s", ip, port)

	log.Print("[INFO] reloading from master ", ip)
	sj = m.loadFromMaster(ip, port)

	if rip := leaderIP(sj.Leader); rip != ip {
		log.Print("[WARN] master changed to ", rip)
		sj = m.loadFromMaster(rip, port)
	}

	return sj, err
}

func (m *Mesos) loadFromMaster(ip string, port string) (sj StateJSON) {
	url := "http://" + ip + ":" + port + "/master/state.json"

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}

	err = json.Unmarshal(body, &sj)
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}

	return sj
}

func (m *Mesos) parseState(sj StateJSON) {
	log.Print("[INFO] Running parseState")

	m.RegisterHosts(sj)
	log.Print("[DEBUG] Done running RegisterHosts")

	for _, fw := range sj.Frameworks {
		for _, task := range fw.Tasks {
			host, err := sj.Followers.hostById(task.FollowerId)
			if err == nil && task.State == "TASK_RUNNING" {
				tname := cleanName(task.Name)

				// Container IP discovery stolen from mesos-dns
				ip := task.IP(host)
				agent := toIP(host)
				log.Printf("[DEBUG] Discovered task %s with ip %s on host %s", tname, ip, agent)

				if task.Resources.Ports != "" && ip == agent {
					for _, port := range yankPorts(task.Resources.Ports) {
						sr := consul.NewRegistration(
							&consulapi.AgentServiceRegistration{
								ID:      fmt.Sprintf("mesos-consul:%s:%s:%d", host, tname, port),
								Name:    tname,
								Port:    port,
								Address: ip,
							},
						)
						m.register(sr)
					}
				} else {
					sr := consul.NewRegistration(
						&consulapi.AgentServiceRegistration{
							ID:      fmt.Sprintf("mesos-consul:%s-%s", host, tname),
							Name:    tname,
							Address: ip,
						},
					)
					sr.Agent = agent
					m.register(sr)
				}

			}
		}
	}

	// Remove completed tasks
	m.deregister()
}

func yankPorts(ports string) []int {
	rhs := strings.Split(ports, "[")[1]
	lhs := strings.Split(rhs, "]")[0]

	yports := []int{}

	mports := strings.Split(lhs, ",")
	for _, mport := range mports {
		pz := strings.Split(strings.TrimSpace(mport), "-")
		lo, _ := strconv.Atoi(pz[0])
		hi, _ := strconv.Atoi(pz[1])

		for t := lo; t <= hi; t++ {
			yports = append(yports, t)
		}
	}

	return yports
}

// Iterate our source options and get the first IP we can
func (t *Task) IP(host string) string {
	var srcs = []string{"docker", "mesos", "host"}

	for _, src := range srcs {
		switch src {
		case "host":
			return toIP(host)
		case "docker", "mesos":
			return t.containerIP(src)
		}
	}
	return ""
}

// Find the IP of a specific container
func (t *Task) containerIP(src string) string {
	ipLabel := ipLabels[src]

	var latestContainerIP string
	var latestTimestamp float64
	for _, status := range t.Statuses {
		if status.State != "TASK_RUNNING" || status.Timestamp <= latestTimestamp {
			continue
		}

		for _, label := range status.Labels {
			if label.Key == ipLabel {
				latestContainerIP = label.Value
				latestTimestamp = status.Timestamp
				break
			}
		}
	}

	return latestContainerIP
}
