BINARY_NAME=MCQPlatform
MAIN_FILE = cmd/main.go
COVERAGE_NAME = coverage

# Check that the architecture is set to either OSX or Linux,
# die with an error otherwise.
# The ARCH variable is to be supplied via the make command.
define check_arch
$(if $(filter $(ARCH), darwin linux), \
	@echo Architecture set to $(ARCH), \
	$(error Please set an ARCH of either "linux" or "darwin"))
endef

build:
	$(call check_arch)
	GOARCH=amd64 GOOS=$(ARCH) go build -o ${BINARY_NAME}-$(ARCH) $(MAIN_FILE)

run:
	./${BINARY_NAME}-${ARCH}

build_and_run: build run

generate:
	go generate ./...

swagger:
	swag fmt
	swag init -d cmd/,pkg/model/cassandra,pkg/model/http,pkg/http/rest -g ../pkg/http/rest/rest.go

clean:
	go clean
	rm -f ${BINARY_NAME}-darwin ${BINARY_NAME}-linux

test:
	go test ./...

test_short:
	go test -short ./...

test_no_cache:
	go test -count=1 ./...

coverage:
	go test ./... -covermode=count -coverprofile ${COVERAGE_NAME}/${COVERAGE_NAME}.out
	go tool cover -html ${COVERAGE_NAME}/${COVERAGE_NAME}.out -o ${COVERAGE_NAME}/${COVERAGE_NAME}.html

coverage_report: coverage
	open ${COVERAGE_NAME}/${COVERAGE_NAME}.html

dep:
	go mod download

.PHONY: build run build_and_run generate swagger clean test test_short test_no_cache coverage coverage_report
