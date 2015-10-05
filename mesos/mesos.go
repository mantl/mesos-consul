package mesos

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/CiscoCloud/mesos-consul/config"
	"github.com/CiscoCloud/mesos-consul/consul"
	"github.com/CiscoCloud/mesos-consul/registry"

	consulapi "github.com/hashicorp/consul/api"
	proto "github.com/mesos/mesos-go/mesosproto"
	"github.com/mesosphere/mesos-dns/records/state"
)

type CacheEntry struct {
	service      *consulapi.AgentServiceRegistration
	isRegistered bool
}

type Mesos struct {
	Registry registry.Registry
	Agents   map[string]string
	Lock     sync.Mutex

	Leader  *proto.MasterInfo
	Masters []*proto.MasterInfo
}

func New(c *config.Config) *Mesos {
	m := new(Mesos)

	if c.Zk == "" {
		return nil
	}

	if consul.IsEnabled() {
		m.Registry = consul.New()
	}

	if m.Registry == nil {
		log.Fatal("[ERROR] No registry specified")
	}

	m.zkDetector(c.Zk)

	return m
}

func (m *Mesos) Refresh() error {
	sj, err := m.loadState()
	if err != nil {
		log.Print("[WARN] loadState failed: ", err.Error())
		return err
	}

	if sj.Leader == "" {
		return errors.New("Empty master")
	}

	if m.Registry.CacheCreate() {
		m.LoadCache()
	}

	m.parseState(sj)

	return nil
}

func (m *Mesos) loadState() (state.State, error) {
	var err error
	var sj state.State

	log.Printf("[DEBUG] loadState() called")

	defer func() {
		if rec := recover(); rec != nil {
			err = errors.New("can't connect to Mesos")
		}
	}()

	mh := m.getLeader()
	if mh.Ip == "" {
		return sj, errors.New("No master in zookeeper")
	}

	log.Printf("[INFO] Zookeeper leader: %s:%s", mh.Ip, mh.PortString)

	log.Print("[INFO] reloading from master ", mh.Ip)
	sj = m.loadFromMaster(mh.Ip, mh.PortString)

	if rip := leaderIP(sj.Leader); rip != mh.Ip {
		log.Print("[WARN] master changed to ", rip)
		sj = m.loadFromMaster(rip, mh.PortString)
	}

	return sj, err
}

func (m *Mesos) loadFromMaster(ip string, port string) (sj state.State) {
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

func (m *Mesos) parseState(sj state.State) {
	log.Print("[INFO] Running parseState")

	m.RegisterHosts(sj)
	log.Print("[DEBUG] Done running RegisterHosts")

	for _, fw := range sj.Frameworks {
		for _, task := range fw.Tasks {
			agent, ok := m.Agents[task.SlaveID]
			if ok && task.State == "TASK_RUNNING" {
				m.registerTask(&task, agent)

			}
		}
	}

	// Remove completed tasks
	m.Registry.Deregister()
}
