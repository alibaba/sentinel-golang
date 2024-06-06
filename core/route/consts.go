package route

const DefaultPort = "80"

const TrafficTagHeader = "x-mse-tag"

const DefaultTag = "base"

const (
	MetadataBaseKey    = "opensergo.io/canary"
	MetadataGrayKey    = "opensergo.io/canary-%s"
	MetadataGrayPrefix = "opensergo.io/canary-"
	MetadataGrayKeyOld = "__micro.service.env__"
)
