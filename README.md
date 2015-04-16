# mesos-consul
Mesos to Consul bridge for service discovery

Service registry bridge for Mesos to Consul

Mesos-consul automatically registers/deregisters services run as Mesos tasks

## Usage

	mesos-consul --registry={registry uri} --zk={zk uri} --refresh={refresh time}

Example

	mesos-consul --registry=consul://127.0.0.1:8500 --zk=zk://127.0.0.1:2181/mesos --refresh=30s

## Todo
  * Add support for tags
  * Automatic registration of Leader/Master/Slave nodes
  * Use task labels for health check data
  * Refactor
