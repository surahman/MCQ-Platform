# Multiple Choice Question Platform

This is a demonstration project in `Golang` that provides an API for Tests and Questionnaires. The tests can be graded
or ungraded with various marking types, please refer to the grading section below.

The limitations on questions and answer options are superficial and are being enforced to demonstrate data validation.

<br/>

**_TODO:_** _There are a number of items that I will complete as time permits:_
-[ ] REST API `score/stats` response pagination: The response to a stats request could contain far too many records to 
     send back in a single request and must be paginated.
-[ ] Redis Cache: To improve the performance of the API, published tests should be cached with an eager write to cache, 
     and lazy loading on cache misses.
-[ ] GraphQL: For performance on bandwidth restricted connections this is essential.

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

## HTTP

Details on the HTTP endpoints can be found in their respective packages below. 

### REST

The HTTP endpoint details are located in the [`http_rest`](pkg/http/rest) package. The model used for REST API calls can
be found in the [`model_http`](pkg/model/http). The REST API server does provide a Swagger UI to examine and test the API
calls and details on request formats.

### GraphQL

Coming soon.

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
