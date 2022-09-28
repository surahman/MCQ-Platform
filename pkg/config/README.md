# Configuration

Configuration loading is designed for containerization in mind. The container engine and orchestrator can mount volumes
(secret or regular) as well as set the environment variables as outlined below.

You may set configurations through both files and environment variables. Please note that environment variables will
override the settings in the configuration files. The configuration files are all expected to be in `YAML` format.

<br/>


## Table of contents

<br/>

## Locations

The configuration loader will search for the configurations in the following order:

| Location                 | Details                                                                                                |
|--------------------------|--------------------------------------------------------------------------------------------------------|
| `/etc/MCQPlatform.conf/` | The `etc` directory is the canonical location for configurations.                                      |
| `$HOME/.MCQPlatform/`    | Configurations can be located in the user's home directory.                                            |
| `.`                      | The root directory where the application is located.                                                   |
| Environment variables    | Finally, the configurations will be loaded from environment variables and override configuration files |

<br/>

## Apache Cassandra

The expected file name is `CassandraConfig.yaml`.

### Configuration File

All the configuration items below are _required_.

| Name                 | Environment Variable Key   | Type         | Description                                                                                   |
|----------------------|----------------------------|--------------|-----------------------------------------------------------------------------------------------|
| **_authentication_** | `CASSANDRA_AUTHENTICATION` |              | **_Parent key for authentication information._**                                              |
| ↳ username           | ↳ `.USERNAME`              | string       | Username for Cassandra session login.                                                         |
| ↳ password           | ↳ `.PASSWORD`              | string       | Password for Cassandra session login.                                                         |
| **_keyspace_**       | `CASSANDRA_KEYSPACE`       |              | **_Parent key for Keyspace configuration._**                                                  |
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

<br/>

## Redis
Redis will be used as an in-memory, fault-tolerant, and replicated cache.

The expected file name is `RedisConfig.yaml`.

<br/>


## Authentication
Configurations in this file will be used to setup all authentication settings (password hashing key etc.).

The expected file name is `AuthenticationConfig.yaml`.

<br/>


## Logging

Structured logging is handled through [Zap](https://pkg.go.dev/go.uber.org/zap).
Please refer to Zap's documentation on how to set the fields below.

Only the `BuiltinConfig` and `BuiltinEncoder` settings can be provided through environment variables that override values
from the configuration files.

**_If specifying settings for the `General Config` or `Encoder Config`, all values must be provided._**

| Name                       | Environment Variable Key      | Type                          | Description                                                                                                 |
|----------------------------|-------------------------------|-------------------------------|-------------------------------------------------------------------------------------------------------------|
| **_BuiltinConfig_**        | `LOGGER_BUILTINCONFIG`        | string                        | Must be one of `Development` or `Production`. Required.                                                     |
| **_BuiltinEncoderConfig_** | `LOGGER_BUILTINENCODERCONFIG` | string                        | Must be one of `Development` or `Production`. Required.                                                     |
| **_General Config_**       | _Not applicable._             | **_General Configurations._** | Please refer to [Zap user documentation](https://pkg.go.dev/go.uber.org/zap#Config).                        |
| ↳ development              |                               | boolean                       |                                                                                                             |
| ↳ disableCaller            |                               | boolean                       |                                                                                                             |
| ↳ disableStacktrace        |                               | boolean                       |                                                                                                             |
| ↳ encoding                 |                               | string                        |                                                                                                             |
| ↳ outputPaths              |                               | string array                  |                                                                                                             |
| ↳ errorOutputPaths         |                               | string array                  |                                                                                                             |
| **_Encoder Config_**       | _Not applicable._             | **_Encoder Configurations._** | Please refer to [Zap user documentation](https://pkg.go.dev/go.uber.org/zap@v1.23.0/zapcore#EncoderConfig). |
| ↳ messageKey               |                               | string                        |
| ↳ levelKey                 |                               | string                        |                                                                                                             |
| ↳ timeKey                  |                               | string                        |                                                                                                             |
| ↳ nameKey                  |                               | string                        |                                                                                                             |
| ↳ callerKey                |                               | string                        |                                                                                                             |
| ↳ functionKey              |                               | string                        |                                                                                                             |
| ↳ stacktraceKey            |                               | string                        |                                                                                                             |
| ↳ skipLineEnding           |                               | boolean                       |                                                                                                             |
| ↳ lineEnding               |                               | string                        |                                                                                                             |
| ↳ consoleSeparator         |                               | string                        |                                                                                                             |

#### Configuration File Format

```yaml
builtin_config: Development | Production
builtin_encoder_config: Development | Production
general_config:
  development: boolean
  disableCaller: boolean
  disableStacktrace: boolean
  encoding: string
  outputPaths: [string, array]
  errorOutputPaths: [string, array]
encoder_config:
  messageKey: string
  levelKey: string
  timeKey: string
  nameKey: string
  callerKey: string
  functionKey: string
  stacktraceKey: string
  skipLineEnding: boolean
  lineEnding: string
  consoleSeparator: string
```

#### Example Environment Variables

```bash
export LOGGER_BUILTINCONFIG="Development"
export LOGGER_BUILTINENCODERCONFIG="Production"
```