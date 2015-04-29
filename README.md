# mesos-consul
Mesos to Consul bridge for service discovery. 

Mesos-consul automatically registers/deregisters services run as Mesos tasks.

This means if you have a Mesos task called `application`, this program will register the application in Consul, and it will be exposed via DNS as `application.service.consul`.

This program also does Mesos leader discovery, so that `leader.mesos.service.consul` will point to the current leader.

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
    "--zk=zk://zookeeper.service.consul:2181/mesos",
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


Usage
-----
### Options
|         Option        | Description |
|-----------------------|-------------|
| `refresh`             | Time between refreshes of Mesos tasks
| `registry-auth`       | The basic authentication username (and optional password), separated by a colon.
| `registry-ssl`        | Use HTTPS while talking to the registry.
| `registry-ssl-verify` | Verify certificates when connecting via SSL.
| `registry-ssl-cert`   | Path to an SSL certificate to use to authenticate to the registry server
| `registry-ssl-cacert` | Path to a CA certificate file, containing one or more CA certificates to use to valid the reigstry server certificate
| `registry-token`      | The registry ACL token
| `zk`*                 | Location of the Mesos path in Zookeeper. The default value is zk://127.0.0.1:2181/mesos


### Consul Registration

##### Leader, Master and Follower Nodes

|    Role    | Registration 
|------------|--------------
| `Leader`   | `leader.mesos.service.consul`, `master.mesos.service.consul`
| `Master`   | `master.mesos.service.consul`
| `Follower` | `follower.mesos.service.consul`

##### Mesos Tasks

Tasks are registered as `task_name.service.consul`

### Todo
  * Add support for tags
  * Use task labels for metadata
  * Support for multiple port tasks
