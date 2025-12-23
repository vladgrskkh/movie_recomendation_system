include .env

# =========================================================================================== #
# HELPERS
# =========================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# =========================================================================================== #
# DEVELOPMENT
# =========================================================================================== #

## run/api: run the image-service application
.PHONY: run/api
run/api:
	go run ./cmd/api

## gen/proto/api: generate protobuf code for the API service (golang)
.PHONY: gen/proto/api
gen/proto/api:
	protoc -I=../proto --go_out=genproto --go_opt=paths=source_relative --go-grpc_out=genproto --go-grpc_opt=paths=source_relative ../proto/common/types.proto ../proto/v1/imageservice/image.proto

# =========================================================================================== #
# QUALITY CONTROL
# =========================================================================================== #
 
## audit: tidy and vendor dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	golangci-lint run
	@echo 'Running tests...'
	go test -race -vet=off ./...
 
## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

# =========================================================================================== #
# BUILD
# =========================================================================================== #
 
current_time := $(shell date -Iseconds)
git_description := $(shell git describe --always --dirty --tags --long)
linker_flags := '-s -X main.buildTime=$(current_time) -X main.version=$(git_description)'

## build/api: build the image service application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'::§
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./cmd/api
