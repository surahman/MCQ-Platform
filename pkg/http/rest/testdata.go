package rest

// configTestData will return a map of test data containing valid and invalid Authorization configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,

		"valid": `
server:
  port_number: 44243
  shutdown_delay: 5
  base_path: api/rest/v1
  swagger_path: /swagger/*any
authorization:
  header_key: Authorization`,

		"out of range port": `
server:
  port_number: 99
  shutdown_delay: 5
  base_path: api/rest/v1
  swagger_path: /swagger/*any
authorization:
  header_key: Authorization`,

		"out of range shutdown delay": `
server:
  port_number: 44243
  shutdown_delay: -1
  base_path: api/rest/v1
  swagger_path: /swagger/*any
authorization:
  header_key: Authorization`,

		"no base path": `
server:
  port_number: 44243
  shutdown_delay: 5
  swagger_path: /swagger/*any
authorization:
  header_key: Authorization`,

		"no swagger path": `
server:
  port_number: 44243
  shutdown_delay: 5
  base_path: api/rest/v1
authorization:
  header_key: Authorization`,

		"no auth header": `
server:
  port_number: 44243
  shutdown_delay: 5
  base_path: api/rest/v1
  swagger_path: /swagger/*any`,
	}
}
