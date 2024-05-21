package route

const defaultPort = "80"

const TrafficTagHeader = "x-mse-tag"

const defaultTag = "base"

const (
	metadataBaseKey    = "opensergo.io/canary"
	metadataGrayKey    = "opensergo.io/canary-%s"
	metadataGrayPrefix = "opensergo.io/canary-"
	metadataGrayKeyOld = "__micro.service.env__"
)
