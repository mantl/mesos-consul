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

type Mesos struct {
	Consul		*consul.Consul
	Masters		*[]MesosHost
	Lock		sync.Mutex
}

func New(c *config.Config, consul *consul.Consul) *Mesos{
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
	
	if (sj.Leader == "") {
		return errors.New("Empty master")
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
				if task.Resources.Ports != "" {
					for _, port := range yankPorts(task.Resources.Ports) {
						m.register(&consulapi.AgentServiceRegistration{
							ID:	fmt.Sprintf("%s:%s:%d",host,tname,port),
							Name:	tname,
							Port:	port,
							Address: toIP(host),
							})
					}
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
	for _,mport := range mports {
		pz := strings.Split(strings.TrimSpace(mport), "-")
		lo, _ := strconv.Atoi(pz[0])
		hi, _ := strconv.Atoi(pz[1])

		for t := lo; t <= hi; t++ {
			yports = append(yports, t)
		}
	}

	return yports
}
