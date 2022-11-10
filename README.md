# Multiple Choice Question Platform

This is a demonstration project in `Golang` that provides an API for Tests and Questionnaires. The tests can be graded
or ungraded with various marking types, please refer to the grading section below.

The limitations on questions and answer options are superficial and are being enforced to demonstrate data validation.
For details on the structure of the JSON requests and responses to the various endpoints, please see the REST API section
below.

<br/>

:warning: **_Transport Layer Security_** :warning:

Encryption is vital to help safeguard against theft of login credentials and JSON Web Tokens.

In a production environment, `TLS` would be the only HTTP protocol over which the API endpoints would be exposed. Setting
up the `TLS`/`SSL` certificated for a Dockerized demonstration environment is unnecessary and complicates the tester's
experience.

Other methods like envelope encryption of payloads add an extra layer of security, but these add an excessive overhead for
the use case and workloads here.

<br/>

:warning: **_Protocols_** :warning:

This demonstration environment will launch both the `HTTP REST` as well as the `GraphQL` over `HTTP` endpoints. This is
unsuitable for a production environment.

Ideally, each of these protocol endpoints would be exposed in its own clusters with auto-scaling, load balancing, and
across availability zones.

<br/>

**_TODO:_** _There are a number of items that I will complete as time permits:_

- [x] REST API `score/stats` response pagination: The response to a stats request could contain far too many records to 
      send back in a single request and must be paginated.
- [x] Denormalize the `responses` table by adding the quiz `Author` with each `response` row. This will allow queries
      requesting statistics to merely make calls to the `response` table. It should then be possible to check
      authorization using the returned rows.
- [x] Redis Cache: To improve the performance of the API, published tests should be cached with an eager write to cache, 
      and lazy loading to the cache on misses.
- [ ] GraphQL: Essential for performance on bandwidth restricted connections.

<br/>

## Logging

Configuration information for the logger can be found in the [`logger`](pkg/logger) package.

<br/>

## Authentication

Information regarding authentication configurations can be found in the [`auth`](pkg/auth) package.

<br/>

## Grading

Responses can be graded or ungraded. Details on the grading schemes can be viewed from the [`grading`](pkg/grading) package.

<br/>

## Cassandra

Information on how to configure the Apache Cassandra connection can be found in the [`cassandra`](pkg/cassandra) package.
Details on the scheme used as well as the design and technical decisions taken can be found in the
[`model_cassandra`](pkg/model/cassandra) package.

A docker-compose file is provided to bring up the necessary services, including a Cassandra single-node cluster. The
keyspace and schema will need to be deployed to the cluster manually. CQL statements to achieve this are located in the
[`model_cassandra`](pkg/model/cassandra) package.

<br/>

## Redis

Information on how to configure the Redis connection can be found in the [`redis`](pkg/redis) package.
Details on the design and technical decisions taken can also be found in the [`redis`](pkg/redis) package.

A docker-compose file is provided to bring up the necessary services, including a minimal Redis three-node cluster. Redis
requires a minimum of three master and three slave nodes for Redis to run in Cluster mode.

<br/>

## HTTP

Details on the HTTP endpoints can be found in their respective packages below. 

### REST

The HTTP endpoint details are located in the [`http_rest`](pkg/http/rest) package. The model used for REST API calls can
be found in the [`model_http`](pkg/model/http).

To review the REST API request and response formats please see the readme in the [`http_handlers`](pkg/http/rest/handlers)
package. The REST API server does also provide a Swagger UI to examine and test the API calls with details on request
formats.

### GraphQL

Work in progress.

<br/>

# Docker Containers

Create containers:
```bash
docker compose up -d
```

Destroy containers:
```bash
docker compose down
```

**Cassandra:**
- Username : `admin`
- Password : `root`
- Keyspace: `mcq_platform`
- 
**Redis:**
- Password : `root`
- Database: `0`, the default database
- Ports: `6379`, `6380`, `6381`, `6382`, `6383`, `6384`
