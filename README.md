# mesos-consul

[![Build Status](https://travis-ci.org/CiscoCloud/mesos-consul.svg)](https://travis-ci.org/CiscoCloud/mesos-consul)

Mesos to Consul bridge for service discovery.

Mesos-consul automatically registers/deregisters services run as Mesos tasks.

This means if you have a Mesos task called `application`, this program will register the application in Consul, and it will be exposed via DNS as `application.service.consul`.

This program also does Mesos leader discovery, so that `leader.mesos.service.consul` will point to the current leader.

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc/generate-toc again -->
**Table of Contents**

- [mesos-consul](#mesos-consul)
    - [Comparisons to other discovery software](#comparisons-to-other-discovery-software)
        - [[Mesos-dns](https://github.com/mesosphere/mesos-dns/)](#mesos-dnshttpsgithubcommesospheremesos-dns)
        - [[Registrator](https://github.com/gliderlabs/registrator)](#registratorhttpsgithubcomgliderlabsregistrator)
    - [Building](#building)
    - [Running](#running)
    - [Usage](#usage)
        - [Options](#options)
        - [Consul Registration](#consul-registration)
            - [Leader, Master and Follower Nodes](#leader-master-and-follower-nodes)
            - [Mesos Tasks](#mesos-tasks)
    - [Todo](#todo)

<!-- markdown-toc end -->

## Comparisons to other discovery software

### [Mesos-dns](https://github.com/mesosphere/mesos-dns/)
This project is similar to mesos-dns in that it polls Mesos to get information about tasks. However, instead of exposing this information via a built-in DNS server, we populate Consul service discovery with this information. Consul then exposes the services via DNS and via its API.

Benefits of using Consul:

* Integration with other tools like [consul-template](https://github.com/hashicorp/consul-template)
* Multi-DC DNS lookups
* Configurable health checks that run on each system

### [Registrator](https://github.com/gliderlabs/registrator)

Registrator is another tool that populates Consul (and other backends like etcd) with the status of Docker containers. However, Registrator is currently limited to reporting on Docker containers and does not track Mesos tasks.


## Building and running mesos-consul

### Building and running in Docker
Mesos-consul can be run in a Docker container via Marathon. 

To build the Docker image, do:
```
docker build -t mesos-consul .
```
To run mesos-consul, start the docker image within Mesos. If your Zookeeper and Marathon services are registered in consul, you can use `.service.consul` to find them, otherwise change the vaules for your environment:
```
curl -X POST -d@mesos-consul.json -H "Content-Type: application/json" http://marathon.service.consul:8080/v2/apps'
```

Where `mesos-consul.json` is similar to (replacing the image with your image):
```
{
  "args": [
    "--zk=zk://zookeeper.service.consul:2181/mesos"
  ],  
  "container": {
    "type": "DOCKER",
    "docker": {
      "network": "BRIDGE",
      "image": "{{ mesos_consul_image }}:{{ mesos_consul_image_tag }}"
    }   
  },  
  "id": "mesos-consul",
  "instances": 1,
  "cpus": 0.1,
  "mem": 256
}
```

You can add options to authenticate via basic http or Consul token.

### Building and running as a service on the Marathon instance

If you don't want to use Docker, it is possible to compile the binary and run it on a Marathon / Mesos master server.

To build it:
* ensure that Go is installed on the build server
* make sure GOPATH is set
* clone this repository and cd into it
* run `make`
* the binary will be created at bin/mesos-consul
* copy this to the Marathon server and start it with `mesos-consul --zk=zk://zookeeper.service.consul:2181/mesos`


## Usage

### Options

|         Option        | Description |
|-----------------------|-------------|
| `version`             | Print mesos-consul version
| `log-level` | Set the Logging level to one of DEBUG, INFO, WARN, ERROR. (default WARN)
| `refresh`             | Time between refreshes of Mesos tasks
| `mesos-ip-order`             | Comma separated list to control the order in which github.com/CiscoCloud/mesos-consul searches or the task IP address. Valid options are 'netinfo', 'mesos', 'docker' and 'host' (default netinfo,mesos,host)
| `healthcheck`             | Enables a http endpoint for health checks. When this flag is enabled, serves health status on 127.0.0.1:24476
| `healthcheck-ip`             | Health check service interface ip
| `healthcheck-port`             | Health check service port. (default 24476)
| `consul-auth`       | The basic authentication username (and optional password), separated by a colon.
| `consul-ssl`        | Use HTTPS while talking to the registry.
| `consul-ssl-verify` | Verify certificates when connecting via SSL.
| `consul-ssl-cert`   | Path to an SSL certificate to use to authenticate to the registry server
| `consul-ssl-cacert` | Path to a CA certificate file, containing one or more CA certificates to use to valid the registry server certificate
| `consul-token`      | The registry ACL token
| `heartbeats-before-remove` | Number of times that registration needs to fail before removing task from Consul. (default: 1)
| `whitelist`         | Only register services matching the provided regex. Can be specified multitple time
| `blacklist`         | Does not register services matching the provided regex. Can be specified multitple time
| `service-name=<name>`      | Service name of the Mesos hosts
| `service-tags=<tag>,...` | Comma delimited list of tags to register the Mesos hosts. Mesos hosts will be registered as (leader|master|follower).<tag>.<service>.service.consul
| `service-id-prefix=<prefix>` | Prefix to use for consul service ids registered by mesos-consul. (default: mesos-consul)
| `task-tag=<pattern:tag>` | Tag tasks matching pattern with given tag. Can be specified multitple times
| `zk`\*                 | Location of the Mesos path in Zookeeper. The default value is zk://127.0.0.1:2181/mesos
| `log-level`            | Level that mesos-consul should log at. Options are [ "DEBUG", "INFO", "WARN", "ERROR" ]. Default is WARN. |
| `group-separator`      | Choose the group separator. Will replace _ in task names (default is empty)


### Consul Registration

#### Leader, Master and Follower Nodes

|    Role    | Registration
|------------|--------------
| `Leader`   | `leader.mesos.service.consul`, `master.mesos.service.consul`
| `Master`   | `master.mesos.service.consul`
| `Follower` | `follower.mesos.service.consul`

#### Mesos Tasks

Tasks are registered as `task_name.service.consul`

#### Tags

Tags can be added to consul by using labels in Mesos. If you are using Marathon you can add a label called `tags` to your service definition with a  comma-separated list of strings that will be registered in consul as tags.

For example, in your marathon service definition:

```
{
  "id": "tagging-test",
  "container": { /*...*/},
  "labels": {
    "tags": "label1,label2,label3"
  }
}
```

This will result in a service `tagging-test` being created in consul with 3 separate tags: `label1` `label2` and `label3`

```
// GET /v1/catalog/service/tagging-test
[
  {
    Node: "consul",
    Address: "10.0.2.15",
    ServiceID: "mesos-consul:10.0.2.15:tagging-test:31562",
    ServiceName: "tagging-test5",
    ServiceTags: [
      "label1",
      "label2",
      "label3"
    ],
    ServiceAddress: "10.0.2.15",
    ServicePort: 31562
  }
]
```
#### Override Task Name

By adding a label `overrideTaskName` with an arbitrary value, the value is used as the service name during consul registration.
Tags are preserved.

## Todo

  * Use task labels for metadata
  * Support for multiple port tasks
