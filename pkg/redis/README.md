# Redis

## Table of contents


<br/>

## Case Study and Justification

To reduce latency and load on the database tier a cache must be used. Redis is ideal for the use case here:
* Highly performant cache.
* Built-in data replication for high-availability across a cache cluster.
* On-disk persistence protects against cold cache scenarios which may see the backend database get overwhelmed with
  requests in the event of a cache failure.
* Automatic failover.
* Keys are evicted using an LRU policy.
* Keys can have an expiration time set via a time-to-live.

Cache Policy:
* Only quizzes will be stored in the cache as they are generally accessible to all users.
* Quizzes will be eager-written to the cache once published. Quizzes that are published are immutable and will not change
  other than to be marked as deleted. An LRU policy, or a TTL if set, will evict keys from the database as necessary.
* Quizzes will be manually evicted from the cache once marked as deleted. This is to ensure data consistency. Deleted
  quizzes will not be loaded into the cache.
* Quizzes will be lazy-read into the cache upon a cache miss. Quizzes that are deleted or not published will not be
  placed into the cache.

:warning: **_Consistency_** :warning:

Redis running in cluster mode is eventually consistent by default for performance reasons. This means that reads, writes,
and deletes might be inconsistent on the same record if executed back-to-back in rapid succession.

Fortunately, in the use case here, records/items placed into the cache are immutable. However, there might be scenarios
where the items placed or removed from the cache might not secure those changes across all the child nodes. In the case
that this causes a cache miss, the data will still be accessible through the backend database. The application must be
tolerant of serving items which are deleted in the backend database.

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

The expected file name is `RedisConfig.yaml`. Unless otherwise specified, all the configuration items below are _required_.

| Name                 | Environment Variable Key | Type         | Description                                                                                                          |
|----------------------|--------------------------|--------------|----------------------------------------------------------------------------------------------------------------------|
| **_Authentication_** | `REDIS_AUTHENTICATION`   |              | **_Parent key for authentication information._**                                                                     |
| ↳ password           | ↳ `.PASSWORD`            | string       | Password for Redis session login.                                                                                    |
| **_Connection_**     | `REDIS_CONNECTION`       |              | **_Parent key for connection configuration._**                                                                       |
| ↳ addrs              | ↳ `.ADDRS`               | string array | An array of cluster IPs to bootstrap the connection. May optionally contain the port numbers.                        |
| ↳ max_redirects      | ↳ `.MAX_REDIRECTS`       | int          | The maximum number of retries before giving up.                                                                      |
| ↳ max_retries        | ↳ `.MAX_RETRIES`         | int          | The maximum number of times to try an operation.                                                                     |
| ↳ pool_size          | ↳ `.POOL_SIZE`           | int          | The connection pool size on a per cluster basis.                                                                     |
| ↳ min_idle_conns     | ↳ `.MIN_IDLE_CONNS`      | int          | The number of minimum idle connections per cluster.                                                                  |
| ↳ read_only          | ↳ `.READ_ONLY`           | bool         | Enables read-only commands on slave nodes. _Optional_                                                                |
| ↳ route_by_latency   | ↳ `.ROUTE_BY_LATENCY`    | bool         | Allows routing read-only commands to the closest master or slave node. It automatically enables ReadOnly. _Optional_ |
| **_Data_**           | `REDIS_DATA`             |              | **_Parent key for data configuration._**                                                                             |
| ↳ ttl                | ↳ `.TTL`                 | int          | The maximum time in seconds tha an item can remain in the cache before it is evicted. _Optional._                    |
