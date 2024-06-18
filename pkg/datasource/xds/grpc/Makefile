all: vet test testrace

build:
	go build github.com/alibaba/sentinel-golang/pkg/datasource/xds/grpc/...

clean:
	go clean -i github.com/alibaba/sentinel-golang/pkg/datasource/xds/grpc/...

deps:
	GO111MODULE=on go get -d -v github.com/alibaba/sentinel-golang/pkg/datasource/xds/grpc/...

proto:
	@ if ! which protoc > /dev/null; then \
		echo "error: protoc not installed" >&2; \
		exit 1; \
	fi
	go generate github.com/alibaba/sentinel-golang/pkg/datasource/xds/grpc/...

test:
	go test -cpu 1,4 -timeout 7m github.com/alibaba/sentinel-golang/pkg/datasource/xds/grpc/...

testsubmodule:
	cd security/advancedtls && go test -cpu 1,4 -timeout 7m github.com/alibaba/sentinel-golang/pkg/datasource/xds/grpc/security/advancedtls/...
	cd security/authorization && go test -cpu 1,4 -timeout 7m github.com/alibaba/sentinel-golang/pkg/datasource/xds/grpc/security/authorization/...

testrace:
	go test -race -cpu 1,4 -timeout 7m github.com/alibaba/sentinel-golang/pkg/datasource/xds/grpc/...

testdeps:
	GO111MODULE=on go get -d -v -t github.com/alibaba/sentinel-golang/pkg/datasource/xds/grpc/...

vet: vetdeps
	./vet.sh

vetdeps:
	./vet.sh -install

.PHONY: \
	all \
	build \
	clean \
	proto \
	test \
	testrace \
	vet \
	vetdeps
