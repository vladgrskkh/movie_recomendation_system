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

## run/api: run the API application
.PHONY: run/api
run/api:
	go run ./cmd/api -db-dsn=${MRS_DB_DSN} -smtp-mailer-api-key=${MAILERSEND_API_KEY} -smtp-sender=${SMTP_USERNAME} -jwt-secret=${JWT_SECRET_KEY}

## run/recommender: run the python grpc service
.PHONY: run/recommender
run/recommender:
	python3 ../predict_service/main.py

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${MRS_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${MRS_DB_DSN} up

## run/docker/api: run the docker container
.PHONY: run/docker/api
run/docker/api:
	docker run -p 8080:8080 vladgrskkh/mrs-api:latest -db-dsn=${MRS_DB_DSN} -smtp-mailer-api-key=${MAILERSEND_API_KEY} -smtp-sender=${SMTP_USERNAME} -jwt-secret=${JWT_SECRET_KEY}

## run/docker/api: run the docker container
.PHONY: run/docker/recommender
run/docker/recommender:
	docker run -p 50051:50051 vladgrskkh/mrs-recommender:latest

## gen/swag: generate swagger documentation
.PHONY: gen/swag
gen/swag:
	swag init -d ./cmd/api -g main.go --parseDependency --output docs/

## gen/proto/api: generate protobuf code for the API service (golang)
.PHONY: gen/proto/api
gen/proto/api:
	protoc -I=../proto --go_out=genproto --go_opt=paths=source_relative --go-grpc_out=genproto --go-grpc_opt=paths=source_relative ../proto/common/types.proto ../proto/v1/predict/predict.proto

## python-env: activate the python virtual environment
.PHONY: python-env
python-env:
	. ../predict_env/bin/activate

## gen/proto/recommender: generate protobuf code for the recommender service (python)
.PHONY: gen/proto/recommender
gen/proto/recommender: python-env
	python3 -m grpc_tools.protoc -I=../proto --python_out=../predict_service --grpc_python_out=../predict_service ../proto/common/types.proto ../proto/v1/predict/predict.proto

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

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./cmd/api

## build/docker/api: build the docker image and push it to docker hub
.PHONY: build/docker/api
build/docker:
	@echo 'Building docker image...'
	docker build --build-arg LINKER_FLAGS=${linker_flags} --tag vladgrskkh/mrs-api .
	@echo 'Pushing docker image...'
	docker push vladgrskkh/mrs-api:latest

## build/docker/recommender: build the docker image and push it to docker hub
.PHONY: build/docker/recommender
build/docker/recommender:
	@echo 'Building docker image...'
	docker build -f ../predict_service/Dockerfile --tag vladgrskkh/mrs-recommender ../predict_service
	@echo 'Pushing docker image...'
	docker push vladgrskkh/mrs-recommender:latest