package config

// CassandraConfigTestData will return a map of test data containing valid and invalid Cassandra configs.
func CassandraConfigTestData() map[string]string {
	testData := make(map[string]string)

	testData["empty"] = ``
	testData["valid"] = `
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
  timeout: 10`
	testData["password_empty"] = `
authentication:
  username: admin
  password:
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: [127.0.0.1]
  proto_version: 4
  timeout: 10`
	testData["username_empty"] = `
authentication:
  username:
  password: root
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: [127.0.0.1]
  proto_version: 4
  timeout: 10`
	testData["keyspace_empty"] = `
authentication:
  username: admin
  password: root
keyspace:
  name:
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: [127.0.0.1]
  proto_version: 4
  timeout: 10`
	testData["consistency_missing"] = `
authentication:
  username: admin
  password: root
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  cluster_ip: [127.0.0.1]
  proto_version: 4
  timeout: 10`
	testData["ip_empty"] = `
authentication:
  username: admin
  password: root
keyspace:
  name: mcq_platform
  replication_class: SimpleStrategy
  replication_factor: 3
connection:
  consistency: quorum
  cluster_ip: []
  proto_version: 4
  timeout: 10`
	testData["timeout_zero"] = `
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
  timeout: 0`

	return testData
}
