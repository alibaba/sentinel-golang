module github.com/alibaba/sentinel-golang

go 1.13

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-00010101000000-000000000000 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.0
	github.com/google/uuid v1.1.1
	github.com/hashicorp/consul/api v1.4.0
	github.com/nacos-group/nacos-sdk-go v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil v2.19.12+incompatible
	github.com/stretchr/testify v1.5.1
	go.uber.org/multierr v1.5.0
	golang.org/x/tools v0.0.0-20200426102838-f3a5411a4c3b // indirect
	google.golang.org/grpc v1.26.0
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/coreos/bbolt v1.3.5 => go.etcd.io/bbolt v1.3.5

replace github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
