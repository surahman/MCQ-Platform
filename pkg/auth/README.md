# Authentication

Configuration loading is designed for containerization in mind. The container engine and orchestrator can mount volumes
(secret or regular) as well as set the environment variables as outlined below.

You may set configurations through both files and environment variables. Please note that environment variables will
override the settings in the configuration files. The configuration files are all expected to be in `YAML` format.

<br/>

## Table of contents


<br/>

### JSON Web Token API Key

API key based authentication is provided through the use of `JWT`s that must be included in the message header section of
an HTTP request:



<br/>

### File Location(s)

The configuration loader will search for the configurations in the following order:

| Location                 | Details                                                                                                |
|--------------------------|--------------------------------------------------------------------------------------------------------|
| `/etc/MCQPlatform.conf/` | The `etc` directory is the canonical location for configurations.                                      |
| `$HOME/.MCQPlatform/`    | Configurations can be located in the user's home directory.                                            |
| `.`                      | The root directory where the application is located.                                                   |
| Environment variables    | Finally, the configurations will be loaded from environment variables and override configuration files |

### Configuration File

The expected file name is `AuthConfig.yaml`. All the configuration items below are _required_.

| Name                 | Environment Variable Key | Type                          | Description                                                                                                          |
|----------------------|--------------------------|-------------------------------|----------------------------------------------------------------------------------------------------------------------|
| **_General_**        | `AUTH_CONFIG `           | **_General Configurations._** | **_Parent key for general authentication configurations._**                                                          |
| ↳ bcryptCost         | ↳ `.BCRYPTCOST`          | int32                         | The [cost](https://pkg.go.dev/golang.org/x/crypto/bcrypt#pkg-constants) value that is used for the BCrypt algorithm. |
| **_JWT_**            | `AUTH_JWT`               | **_JWT Configurations._**     | **_Parent key for JSON Web Token configurations._**                                                                  |
| ↳ key                | ↳ `.KEY`                 | string                        | The encryption key used for the JSON Web Token.                                                                      |
| ↳ expirationDuration | ↳ `.EXPIRATIONDURATION`  | int32                         | The validity duration in seconds for the JSON Web Token.                                                             |

#### Example Configuration File

```yaml
general:
  bcryptCost: 8
jwt:
  key: some-long-random-key
  expirationDuration: 600
```

#### Example Environment Variables

```bash
export AUTH_CONFIG.BCRYPTCOST=8
export AUTH_JWT.KEY="some-long-random-key"
```
