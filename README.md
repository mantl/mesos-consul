# mesos-consul
Mesos to Consul bridge for service discovery

Service registry bridge for Mesos to Consul

Mesos-consul automatically registers/deregisters services run as Mesos tasks

Usage
-----
### Options
|         Option        | Description |
|-----------------------|-------------|
| `refresh`             | Time between refreshes of Mesos tasks
| `registry`*           | Location of the registry instance. The default value is consul://127.0.0.1:8500
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
  * Automatic registration of Master nodes
  * Use task labels for metadata
  * Support for multiple port tasks
