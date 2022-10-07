package auth

// configTestData will return a map of test data containing valid and invalid Authorization configs.
func configTestData() map[string]string {
	testData := make(map[string]string)

	testData["empty"] = ``

	testData["valid"] = `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  expiration_duration: 600
general:
  bcrypt_cost: 8`

	testData["bcrypt_cost_below_4"] = `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  expiration_duration: 600
general:
  bcrypt_cost: 2`

	testData["bcrypt_cost_below_31"] = `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  expiration_duration: 600
general:
  bcrypt_cost: 32`

	testData["jwt_expiration_below_10s"] = `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9U
  expiration_duration: 5
general:
  bcrypt_cost: 8`

	testData["jwt_key_below_8"] = `
jwt:
  key: kYzJdnp
  expiration_duration: 600
general:
  bcrypt_cost: 8`

	testData["jwt_key_above_256"] = `
jwt:
  key: kYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9UkYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9UkYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9UkYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5tcxrVmeQq1gA4mUfzYQ9t9UkYzJdnpm6Lj2E7AobZ35RE2itZ2ws82U5
  expiration_duration: 600
general:
  bcrypt_cost: 8`

	return testData
}
