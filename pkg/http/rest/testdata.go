package rest

// configTestData will return a map of test data containing valid and invalid Authorization configs.
func configTestData() map[string]string {
	return map[string]string{
		"empty": ``,
		"valid": `
general:
  port_number: 44243
  shutdown_delay: 5
  base_path: api/rest/v1
  swagger_path: /swagger/*any`,

		"out of range port": `
general:
  port_number: 99
  shutdown_delay: 5
  base_path: api/rest/v1
  swagger_path: /swagger/*any`,

		"out of range shutdown delay": `
general:
  port_number: 44243
  shutdown_delay: -1
  base_path: api/rest/v1
  swagger_path: /swagger/*any`,

		"no base path": `
general:
  port_number: 44243
  shutdown_delay: 5
  swagger_path: /swagger/*any`,

		"no swagger path": `
general:
  port_number: 44243
  shutdown_delay: 5
  base_path: api/rest/v1`,
	}
}
