package route

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"strings"
)

func FilterServiceInstancesWithTag(ctx context.Context, service *model.Service) {
	defer func() {
		if service != nil {
			fmt.Printf("[GetServiceInfoAfter] service instances: %v\n", service.Hosts)
		}
	}()

	if service == nil || service.Hosts == nil || len(service.Hosts) == 0 {
		return
	}

	// 按流量标签过滤
	trafficTag := getTrafficTag(ctx)
	fmt.Printf("[GetServiceInfoAfter] trafficTag: %s\n", trafficTag)
	if trafficTag != "" {
		tagInstances, baseInstances := instanceFilter(service.Hosts, trafficTag)
		fmt.Printf("[GetServiceInfoAfter] filter by traffic tag, tag instance: %v, base instance: %v\n", tagInstances, baseInstances)
		if trafficTag != defaultTag && len(tagInstances) != 0 {
			service.Hosts = tagInstances
			return
		}
		if len(baseInstances) != 0 {
			service.Hosts = baseInstances
		}
		return
	}

	// 无流量标签, 按灰度标签过滤
	podTag := getPodTag(ctx)
	fmt.Printf("[GetServiceInfoAfter] podTag: %s\n", podTag)
	tagInstances, baseInstances := instanceFilter(service.Hosts, podTag)
	fmt.Printf("[GetServiceInfoAfter] filter by pod tag, tag instance: %v, base instance: %v\n", tagInstances, baseInstances)
	if podTag != "" && podTag != defaultTag { // 灰度节点
		if len(tagInstances) != 0 {
			service.Hosts = tagInstances
			setTrafficTag(ctx, podTag) // 流量标签在baggage中更新为灰度节点标签
			return
		}
	}
	if len(baseInstances) != 0 {
		service.Hosts = baseInstances
	}
}

func instanceFilter(instances []model.Instance, tag string) ([]model.Instance, []model.Instance) {
	tagInstances, baseInstances := make([]model.Instance, 0), make([]model.Instance, 0)
	for _, instance := range instances {
		if instance.Metadata == nil || len(instance.Metadata) == 0 {
			baseInstances = append(baseInstances, instance)
			continue
		}

		if v, ok := instance.Metadata[fmt.Sprintf(metadataGrayKey, tag)]; ok && v == tag { //新灰度标签
			tagInstances = append(tagInstances, instance)
			continue
		}

		var hasOldGrayKey bool
		if v, ok := instance.Metadata[metadataGrayKeyOld]; ok { //老灰度标签
			hasOldGrayKey = true
			var values []map[string]interface{}
			err := json.Unmarshal([]byte(v), &values)
			if err == nil {
				for _, value := range values {
					if value["tag"] == tag {
						tagInstances = append(tagInstances, instance)
						continue
					}
				}
			}
		}

		if _, ok := instance.Metadata[metadataBaseKey]; ok {
			baseInstances = append(baseInstances, instance)
			continue
		}

		if !hasOldGrayKey { //但是他有新灰度标签,需要修复,是其他类型的灰度,没有匹配到当前灰度
			var grayInstance bool
			for k, _ := range instance.Metadata {
				if strings.Contains(k, metadataGrayPrefix) {
					grayInstance = true
					break
				}
			}

			if grayInstance {
				continue
			}
			baseInstances = append(baseInstances, instance)
		}
	}

	if tag == "" || tag == defaultTag {
		tagInstances = baseInstances
	}
	return tagInstances, baseInstances
}

func AddTagMetadata(ctx context.Context, instance *model.Instance) {
	if instance == nil {
		return
	}

	var metadataKey, metadataValue string

	if podTag := getPodTag(ctx); podTag != "" && podTag != defaultTag {
		metadataKey = fmt.Sprintf(metadataGrayKey, podTag)
		metadataValue = podTag
	} else {
		metadataKey = metadataBaseKey
		metadataValue = ""
	}

	if instance.Metadata == nil {
		instance.Metadata = make(map[string]string)
	}

	instance.Metadata[metadataKey] = metadataValue

}
