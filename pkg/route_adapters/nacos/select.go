package nacos

import (
	"context"
	sentinelroute "github.com/alibaba/sentinel-golang/core/route/nacos"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
)

func FilterServiceInstances(service *model.Service) {
	sentinelroute.FilterServiceInstancesWithTag(context.Background(), service)
}
