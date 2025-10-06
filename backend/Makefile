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
	go run ./cmd/api