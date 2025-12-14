## gen/proto/api: generate protobuf code for the API service (golang)
.PHONY: gen/proto/api
gen/proto/api:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./common/types.proto ./v1/imageservice/image.proto ./v1/predict/predict.proto