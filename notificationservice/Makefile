include .envrc

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

## run/notification-service: run the notification-service application
.PHONY: run/notification-service
run/api:
	go run ./cmd/api

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
 
## build/api: build the cmd/api application
.PHONY: build/notification-service
build/notification-service:
	@echo 'Building cmd/api...'
	CGO_ENABLED=1 go build -o=./bin/api ./cmd/api

## build/docker: build the docker image and push it to docker hub
.PHONY: build/docker
build/docker:
	@echo 'Building docker image...'
	docker build --tag vladgrskkh/mrs-notification-service .
	@echo 'Pushing docker image...'
	docker push vladgrskkh/mrs-notification-service:latest
