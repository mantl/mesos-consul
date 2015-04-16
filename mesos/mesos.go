package mesos

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/CiscoCloud/mesos-consul/registry"

)

func (l *MesosLeader) Refresh(r registry.RegistryAdapter) error {
	sj, err := l.findMaster()
	if err != nil {
		log.Print("No master")
		return err
	}
	
	if (sj.Leader == "") {
		log.Print("Unexpected error")
		return errors.New("Empty master")
	}

	l.parseState(sj, r)

	return nil
}

func (l *MesosLeader) findMaster() (StateJSON, error) {
	var sj StateJSON

	if ip, port := l.getLeader(); ip != "" {
		log.Print("Zookeeper says the leader is: ", ip)

		sj, _ = l.loadWrap(ip, port)
		if (sj.Leader == "") {
			log.Print("Warning: Zookeeper is wrong about leader")
			return sj, errors.New("no master")
		} else {
			return sj, nil
		}
	}

	return sj, errors.New("no master")
}

func (l *MesosLeader) loadWrap(ip string, port string) (StateJSON, error) {
	var err error
	var sj StateJSON

	defer func() {
		if rec := recover(); rec != nil {
			err = errors.New("can't connect to Mesos")
		}
	}()

	log.Print("reloading from master ", ip)
	sj = l.loadFromMaster(ip, port)

	if rip := leaderIP(sj.Leader); rip != ip {
		log.Print("Warning: master changed to ", rip)
		sj = l.loadFromMaster(rip, port)
	}

	return sj, err
}

func (l *MesosLeader) loadFromMaster(ip string, port string) (sj StateJSON) {
	url := "http://" + ip + ":" + port + "/master/state.json"

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(body, &sj)
	if err != nil {
		log.Fatal(err)
	}

	return sj
}

func (l *MesosLeader) getLeader() (string, string) {
	l.leaderLock.Lock()
	defer l.leaderLock.Unlock()
	return l.host, l.port
}

func (l *MesosLeader) parseState(sj StateJSON, r registry.RegistryAdapter) {
	log.Print("Running parseState")
	for _, fw := range sj.Frameworks {
		fname := cleanName(fw.Name)
		for _, task := range fw.Tasks {
			host, err := l.hostBySlaveId(sj.Slaves, task.SlaveId)
			if err == nil && task.State == "TASK_RUNNING" {
				tname := cleanName(task.Name)
				if task.Resources.Ports != "" {
					for _, port := range yankPorts(task.Resources.Ports) {
						s := new(registry.Service)
						s.ID = fmt.Sprintf("%s:%s:%d", host, tname, port)
						s.Name = tname
						s.Port = port
						ip, err := net.LookupIP(host)
						if err != nil {
							s.IP = host
						} else {
							s.IP = ip[0].String()
						}

						l.register(r, s)
					}
				}
			}
		}
	}

	// Remove completed tasks
	l.deregister(r)
}

func (l *MesosLeader) hostBySlaveId(slist Slaves, slaveId string) (string, error) {
	for _, s := range slist {
		if s.Id == slaveId {
			return s.Hostname, nil
		}
	}

	return "", errors.New("not found")
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
