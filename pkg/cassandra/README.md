# Apache Cassandra

Configuration loading is designed for containerization in mind. The container engine and orchestrator can mount volumes
(secret or regular) as well as set the environment variables as outlined below.

You may set configurations through both files and environment variables. Please note that environment variables will
override the settings in the configuration files. The configuration files are all expected to be in `YAML` format.

<br/>

## Table of contents

- [File Location(s)](#file-locations)
- [Configuration File](#configuration-file)
    - [Example Configuration File](#example-configuration-file)
    - [Example Environment Variables](#example-environment-variables)
- [Design Documentation](#design-documentation)

<br/>

### File Location(s)

The configuration loader will search for the configurations in the following order:

| Location                 | Details                                                                                                |
|--------------------------|--------------------------------------------------------------------------------------------------------|
| `/etc/MCQPlatform.conf/` | The `etc` directory is the canonical location for configurations.                                      |
| `$HOME/.MCQPlatform/`    | Configurations can be located in the user's home directory.                                            |
| ./configs/`              | The config folder in the root directory where the application is located.                              |
| Environment variables    | Finally, the configurations will be loaded from environment variables and override configuration files |

### Configuration File

The expected file name is `CassandraConfig.yaml`. All the configuration items below are _required_.

| Name                 | Environment Variable Key   | Type         | Description                                                                                   |
|----------------------|----------------------------|--------------|-----------------------------------------------------------------------------------------------|
| **_Authentication_** | `CASSANDRA_AUTHENTICATION` |              | **_Parent key for authentication information._**                                              |
| ↳ username           | ↳ `.USERNAME`              | string       | Username for Cassandra session login.                                                         |
| ↳ password           | ↳ `.PASSWORD`              | string       | Password for Cassandra session login.                                                         |
| **_Keyspace_**       | `CASSANDRA_KEYSPACE`       |              | **_Parent key for Keyspace configuration._**                                                  |
| ↳ name               | ↳ `.NAME`                  | string       | Keyspace name.                                                                                |
| ↳ replication_class  | ↳ `.REPLICATION_CLASS`     | string       | Data replication strategy.                                                                    |
| ↳ replication_factor | ↳ `.REPLICATION_FACTOR`    | int          | Data replication factor.                                                                      |
| **_Connection_**     | `CASSANDRA_CONNECTION`     |              | **_Parent key for connection configuration._**                                                |
| ↳ consistency        | ↳ `.CONSISTENCY`           | string       | Keyspace name.                                                                                |
| ↳ cluster_ip         | ↳ `.CLUSTER_IP`            | string array | An array of cluster IPs to bootstrap the connection. May optionally contain the port numbers. |
| ↳ proto_version      | ↳ `.PROTO_VERSION`         | int          | Connection protocol version.                                                                  |
| ↳ timeout            | ↳ `.TIMEOUT`               | int          | Connection timeout in seconds.                                                                |

:warning: **_Consistency_** :warning:

It is highly advised that the consistency level be set to `Quorum`. `Quorum` provides a good balance between read and
write performance thereby providing an appropriate balance between consistency and availability. Setting the level to
`All` can negatively impact read and write performance.

A balance between consistency and availability is vital to mitigate data inconsistencies with user account credentials.

#### Example Configuration File

```yaml
authentication:
  username: admin
  password: root
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: [127.0.0.1]
  proto_version: 4
  timeout: 10
```

#### Example Environment Variables

```bash
export CASSANDRA_AUTHENTICATION.USERNAME=admin
export CASSANDRA_AUTHENTICATION.PASSWORD=root
```

### Design Documentation
The technology selection case study and table schema design documentation can be found [here](../model/cassandra).
