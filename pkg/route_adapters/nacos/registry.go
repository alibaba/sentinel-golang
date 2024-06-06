package nacos

import (
	"context"
	sentinelroute "github.com/alibaba/sentinel-golang/core/route/nacos"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
)

func AddTagMetadata(instance *model.Instance) {
	sentinelroute.AddTagMetadata(context.Background(), instance)
}

func BatchAddTagMetadata(instances []model.Instance) []model.Instance {
	if instances == nil || len(instances) == 0 {
		return instances
	}

	for _, instance := range instances {
		sentinelroute.AddTagMetadata(context.Background(), &instance)
	}

	return instances
}
