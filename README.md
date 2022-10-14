# Multiple Choice Question Platform


<br/>

## Logging

Configuration information for the logger can be found in the [`logger`](pkg/logger) package.

<br/>

## Authentication

Information regarding authentication configurations can be found in the [`auth`](pkg/auth) package.

<br/>

## Cassandra

Information on how to configure the Apache Cassandra connection can be found in the [`cassandra`](pkg/cassandra) package.
Details on the scheme used as well as the design and technical decisions taken can be found in the
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

## Grading

Responses can be graded or ungraded. Details on the grading schemes can be viewed from the [`grading`](pkg/grading) package.

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
