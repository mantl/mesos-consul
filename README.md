# mesos-consul

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

## Building
```
docker build -t mesos-consul . 
```

## Running
Mesos-consul can be run in a Docker container via Marathon. If your Zookeeper and Marathon services are registered in consul, you can use `.service.consul` to find them, otherwise change the vaules for your environment:


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


## Usage

### Options

|         Option        | Description |
|-----------------------|-------------|
| `refresh`             | Time between refreshes of Mesos tasks
| `consul-auth`       | The basic authentication username (and optional password), separated by a colon.
| `consul-ssl`        | Use HTTPS while talking to the registry.
| `consul-ssl-verify` | Verify certificates when connecting via SSL.
| `consul-ssl-cert`   | Path to an SSL certificate to use to authenticate to the registry server
| `consul-ssl-cacert` | Path to a CA certificate file, containing one or more CA certificates to use to valid the reigstry server certificate
| `consul-token`      | The registry ACL token
| `zk`*                 | Location of the Mesos path in Zookeeper. The default value is zk://127.0.0.1:2181/mesos


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

## Todo

  * Use task labels for metadata
  * Support for multiple port tasks
