module github.com/alibaba/sentinel-golang

go 1.16

replace google.golang.org/grpc => github.com/EndlessSeeker/grpc-go v1.57.1

require (
	dubbo.apache.org/dubbo-go/v3 v3.1.1
	github.com/dubbogo/gost v1.14.0
	github.com/envoyproxy/go-control-plane v0.12.0
	github.com/fsnotify/fsnotify v1.6.0
	github.com/golang/protobuf v1.5.3
	github.com/google/uuid v1.4.0
	github.com/lestrrat-go/jwx/v2 v2.0.21
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.13.0
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/shirou/gopsutil/v3 v3.22.2
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/otel v1.24.0
	go.uber.org/multierr v1.11.0
	google.golang.org/genproto/googleapis/api v0.0.0-20240102182953-50ed04b92917 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240102182953-50ed04b92917 // indirect
	google.golang.org/grpc v1.61.1
	google.golang.org/protobuf v1.32.0
	gopkg.in/yaml.v2 v2.4.0
)
