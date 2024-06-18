package bootstrap

import (
	"errors"
	"fmt"
	v3core "github.com/alibaba/sentinel-golang/pkg/datasource/xds/go-control-plane/envoy/config/core/v3"
	"github.com/alibaba/sentinel-golang/pkg/datasource/xds/utils"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
	"strings"
)

func InitNode() (*v3core.Node, error) {
	podIP := os.Getenv(utils.EnvPodIP)
	if podIP == "" {
		return nil, errors.New("[InitNode] KUBERNETES_POD_IP is not set in env")
	}
	podName := os.Getenv(utils.ENVPodName)
	if podName == "" {
		return nil, errors.New("[InitNode] KUBERNETES_POD_NAME is not set in env")
	}
	namespace := os.Getenv(utils.EnvNamespace)
	if namespace == "" {
		return nil, errors.New("[InitNode] KUBERNETES_POD_NAMESPACE is not set in env")
	}
	clusterDomain := os.Getenv(utils.EnvClusterDomain)
	if clusterDomain == "" {
		clusterDomain = utils.DefaultClusterDomain
	}
	nodeID := genNodeID(podIP, podName, namespace, clusterDomain)
	metadata := parseMetaEnvs(os.Getenv(utils.EnvXdsMetas), os.Getenv(utils.EnvIstioVersion), podIP)

	return &v3core.Node{
		Id:       nodeID,
		Metadata: metadata,
	}, nil
}

func genNodeID(podIP, podName, namespace, clusterDomain string) string {
	//"sidecar~" + podIP + "~" + podName + "." + namespace + "~" + namespace + ".svc." + domain,
	return fmt.Sprintf("sidecar~%s~%s.%s~%s.svc.%s", podIP, podName, namespace, namespace, clusterDomain)
}

func parseMetaEnvs(envs, istioVersion, podIP string) *structpb.Struct {
	defaultMeta := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"ISTIO_VERSION": {
				Kind: &structpb.Value_StringValue{StringValue: istioVersion},
			},
		},
	}
	if len(envs) == 0 {
		return defaultMeta
	}

	pbMeta := &structpb.Struct{
		Fields: map[string]*structpb.Value{},
	}
	err := protojson.Unmarshal([]byte(envs), pbMeta)
	if err != nil {
		return defaultMeta
	}
	if ips, ok := pbMeta.Fields["INSTANCE_IPS"]; ok {
		existIPs := ips.GetStringValue()
		if existIPs == "" {
			existIPs = podIP
		} else if !strings.Contains(existIPs, podIP) {
			existIPs = existIPs + "," + podIP
		}
		pbMeta.Fields["INSTANCE_IPS"] = &structpb.Value{
			Kind: &structpb.Value_StringValue{StringValue: existIPs},
		}
	}
	return pbMeta
}
