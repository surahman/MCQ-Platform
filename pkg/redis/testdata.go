package redis

// configTestData will return a map of test data containing valid and invalid Cassandra configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"valid": `
authentication:
  password: root
connection:
  addrs: [127.0.0.1:6379, 127.0.0.1:6380, 127.0.0.1:6381, 127.0.0.1:6382, 127.0.0.1:6383, 127.0.0.1:6384]
  max_connection_attempts: 5
  max_redirects: 3
  max_retries: 3
  pool_size: 4
  min_idle_conns: 1
  read_only: false
  route_by_latency: false
data:
  ttl: 3600`,

		"password_empty": `
authentication:
  password:
connection:
  addrs: [127.0.0.1:6379, 127.0.0.1:6380, 127.0.0.1:6381, 127.0.0.1:6382, 127.0.0.1:6383, 127.0.0.1:6384]
  max_connection_attempts: 5
  max_redirects: 3
  max_retries: 3
  pool_size: 4
  min_idle_conns: 1
  read_only: false
  route_by_latency: false
data:
  ttl: 3600`,

		"no_addrs": `
authentication:
  password: root
connection:
  addrs: []
  max_connection_attempts: 5
  max_redirects: 3
  max_retries: 3
  pool_size: 4
  min_idle_conns: 1
  read_only: false
  route_by_latency: false
data:
  ttl: 3600`,

		"invalid_max_redirects": `
authentication:
  password: root
connection:
  addrs: [127.0.0.1:6379, 127.0.0.1:6380, 127.0.0.1:6381, 127.0.0.1:6382, 127.0.0.1:6383, 127.0.0.1:6384]
  max_connection_attempts: 5
  max_redirects: 0
  max_retries: 3
  pool_size: 4
  min_idle_conns: 1
  read_only: false
  route_by_latency: false
data:
  ttl: 3600`,

		"invalid_max_retries": `
authentication:
  password: root
connection:
  addrs: [127.0.0.1:6379, 127.0.0.1:6380, 127.0.0.1:6381, 127.0.0.1:6382, 127.0.0.1:6383, 127.0.0.1:6384]
  max_connection_attempts: 5
  max_redirects: 3
  max_retries: 0
  pool_size: 4
  min_idle_conns: 1
  read_only: false
  route_by_latency: false
data:
  ttl: 3600`,

		"invalid_pool_size": `
authentication:
  password: root
connection:
  addrs: [127.0.0.1:6379, 127.0.0.1:6380, 127.0.0.1:6381, 127.0.0.1:6382, 127.0.0.1:6383, 127.0.0.1:6384]
  max_connection_attempts: 5
  max_redirects: 3
  max_retries: 3
  pool_size: 0
  min_idle_conns: 1
  read_only: false
  route_by_latency: false
data:
  ttl: 3600`,

		"invalid_min_idle_conns": `
authentication:
  password: root
connection:
  addrs: [127.0.0.1:6379, 127.0.0.1:6380, 127.0.0.1:6381, 127.0.0.1:6382, 127.0.0.1:6383, 127.0.0.1:6384]
  max_connection_attempts: 5
  max_redirects: 3
  max_retries: 3
  pool_size: 4
  min_idle_conns: 0
  read_only: false
  route_by_latency: false
data:
  ttl: 3600`,

		"invalid_min_ttl": `
authentication:
  password: root
connection:
  addrs: [127.0.0.1:6379, 127.0.0.1:6380, 127.0.0.1:6381, 127.0.0.1:6382, 127.0.0.1:6383, 127.0.0.1:6384]
  max_connection_attempts: 5
  max_redirects: 3
  max_retries: 3
  pool_size: 4
  min_idle_conns: 1
  read_only: false
  route_by_latency: false
data:
  ttl: 59`,

		"invalid_min_max_conn_attempts": `
authentication:
  password: root
connection:
  addrs: [127.0.0.1:6379, 127.0.0.1:6380, 127.0.0.1:6381, 127.0.0.1:6382, 127.0.0.1:6383, 127.0.0.1:6384]
  max_connection_attempts: 0
  max_redirects: 3
  max_retries: 3
  pool_size: 4
  min_idle_conns: 1
  read_only: false
  route_by_latency: false
data:
  ttl: 3600`,

		"test_suite": `
authentication:
  password: root
connection:
  addrs: [127.0.0.1:6379, 127.0.0.1:6380, 127.0.0.1:6381, 127.0.0.1:6382, 127.0.0.1:6383, 127.0.0.1:6384]
  max_connection_attempts: 2
  max_redirects: 3
  max_retries: 3
  pool_size: 4
  min_idle_conns: 1
  read_only: false
  route_by_latency: false
data:
  ttl: 3600`,
	}
}
