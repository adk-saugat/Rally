.PHONY: gen clean tidy

export PATH := $(PATH):$(shell go env GOPATH)/bin

gen:
	mkdir -p gen
	protoc \
	  -I proto \
	  --go_out=gen --go_opt=paths=source_relative \
	  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
	  auth/v1/auth.proto \
	  group/v1/group.proto \
	  location/v1/location.proto

tidy:
	go mod tidy

clean:
	rm -rf gen/auth gen/group gen/location
