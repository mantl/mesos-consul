package mesos

type MesosHost struct {
	Ip           string
	Host         string
	Port         int
	PortString   string
	IsLeader     bool
	IsRegistered bool
}
