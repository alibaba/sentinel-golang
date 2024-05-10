package route

const defaultPort = "80"

const TrafficTagHeader = "x-mse-gray-tag"

const defaultTag = "base"

const (
	EnvPodGrayTag      = "MSE_ALICLOUD_SERVICE_TAG"
	BaggageGrayTag     = "x-mse-gray-tag"
	BaseVersion        = "base"
	MetadataBaseKey    = "opensergo.io/canary"
	MetadataGrayKey    = "opensergo.io/canary-%s"
	MetadataGrayPrefix = "opensergo.io/canary-"
	MetadataGrayKeyOld = "__micro.service.env__"
)
