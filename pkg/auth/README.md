# Authentication

Configuration loading is designed for containerization in mind. The container engine and orchestrator can mount volumes
(secret or regular) as well as set the environment variables as outlined below.

You may set configurations through both files and environment variables. Please note that environment variables will
override the settings in the configuration files. The configuration files are all expected to be in `YAML` format.

<br/>

## Table of contents

- [JSON Web Token API Key](#json-web-token-api-key)
- [File Location(s)](#file-locations)
- [Configuration File](#configuration-file)
    - [Example Configuration File](#example-configuration-file)
    - [Example Environment Variables](#example-environment-variables)

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
| `./configs/`             | The config folder in the root directory where the application is located.                              |
| Environment variables    | Finally, the configurations will be loaded from environment variables and override configuration files |

### Configuration File

The expected file name is `AuthConfig.yaml`. All the configuration items below are _required_.

| Name                  | Environment Variable Key | Type                          | Description                                                                                                          |
|-----------------------|--------------------------|-------------------------------|----------------------------------------------------------------------------------------------------------------------|
| **_JWT_**             | `AUTH_JWT`               | **_JWT Configurations._**     | **_Parent key for JSON Web Token configurations._**                                                                  |
| ↳ key                 | ↳ `.KEY`                 | string                        | The encryption key used for the JSON Web Token.                                                                      |
| ↳ issuer              | ↳ `.ISSUER`              | string                        | The issuer of the JSON Web Token.                                                                                    |
| ↳ expiration_duration | ↳ `.EXPIRATION_DURATION` | int64                         | The validity duration in seconds for the JSON Web Token.                                                             |
| ↳ refresh_threshold   | ↳ `.REFRESH_THRESHOLD`   | int64                         | The seconds before expiration that a JSON Web Token can be refreshed before.                                         |
| **_General_**         | `AUTH_CONFIG `           | **_General Configurations._** | **_Parent key for general authentication configurations._**                                                          |
| ↳ bcrypt_cost         | ↳ `.BCRYPT_COST`         | int                           | The [cost](https://pkg.go.dev/golang.org/x/crypto/bcrypt#pkg-constants) value that is used for the BCrypt algorithm. |
| ↳ crypto_secret       | ↳ `.CRYPTO_SECRET`       | string                        | A 32 character secret key to be used for AES256 encryption and decryption.                                           |

#### Example Configuration File

```yaml
jwt:
  key: some-long-random-key
  issuer: issuer of the token
  expiration_duration: 600
  refresh_threshold: 60
general:
  bcrypt_cost: 8
```

#### Example Environment Variables

```bash
export AUTH_CONFIG.BCRYPT_COST=8
export AUTH_JWT.KEY="some-long-random-key"
```
