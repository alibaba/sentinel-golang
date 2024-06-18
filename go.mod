module github.com/alibaba/sentinel-golang

go 1.16

require (
	github.com/alibaba/sentinel-golang/pkg/datasource/xds v0.0.0-00010101000000-000000000000
	github.com/fsnotify/fsnotify v1.6.0
	github.com/google/uuid v1.6.0
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/nacos-group/nacos-sdk-go/v2 v2.2.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.13.0
	github.com/shirou/gopsutil/v3 v3.22.2
	github.com/stretchr/testify v1.9.0
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.6.0
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/alibaba/sentinel-golang/pkg/datasource/xds => ./pkg/datasource/xds
