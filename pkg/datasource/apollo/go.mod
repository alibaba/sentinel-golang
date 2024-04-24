module github.com/alibaba/sentinel-golang/pkg/datasource/apollo

go 1.13

replace github.com/alibaba/sentinel-golang => ../../../

require (
	github.com/alibaba/sentinel-golang v1.0.4
	github.com/apolloconfig/agollo/v4 v4.0.9
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.9.0
)

require github.com/spf13/viper v1.9.0 // indirect
