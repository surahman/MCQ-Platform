package graphql

// configTestData will return a map of test data containing valid and invalid Authorization configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"valid": `
server:
  port_number: 44255
  shutdown_delay: 5
  base_path: api/graphql/v1
  playground_path: /playground
authorization:
  header_key: Authorization`,

		"out of range port": `
server:
  port_number: 99
  shutdown_delay: 5
  base_path: api/graphql/v1
  playground_path: /playground
authorization:
  header_key: Authorization`,

		"out of range shutdown delay": `
server:
  port_number: 44255
  shutdown_delay: -1
  base_path: api/graphql/v1
  playground_path: /playground
authorization:
  header_key: Authorization`,

		"no base path": `
server:
  port_number: 44243
  shutdown_delay: 5
  playground_path: /playground
authorization:
  header_key: Authorization`,

		"no swagger path": `
server:
  port_number: 44255
  shutdown_delay: 5
  base_path: api/graphql/v1
authorization:
  header_key: Authorization`,

		"no auth header": `
server:
  port_number: 44255
  shutdown_delay: 5
  base_path: api/graphql/v1
  playground_path: /playground`,
	}
}
