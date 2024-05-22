package fallback

type TargetResourceType int64

const (
	RpcResourceType TargetResourceType = 1
	WebResourceType TargetResourceType = 2
)

type FunctionType int64

const (
	FlowType    FunctionType = 1
	Isolation   FunctionType = 6
	HotspotWeb  FunctionType = 4
	HotspotHttp FunctionType = 11
)

type Rule struct {
	TargetResourceType TargetResourceType        `json:"targetResourceType"`
	TargetMap          map[string][]FunctionType `json:"targetMap"`
	FallbackBehavior   []byte                    `json:"fallbackBehavior"`
}

type WebBlockFallbackBehavior struct {
	WebFallbackMode    int64  `json:"webFallbackMode"`
	WebRespStatusCode  int64  `json:"webRespStatusCode"`
	WebRespMessage     string `json:"webRespMessage"`
	WebRespContentType int64  `json:"webRespContentType"`
	WebRedirectUrl     string `json:"webRedirectUrl"`
}

type RpcBlockFallbackBehavior struct {
	RpcFallbackMode             int64  `json:"rpcFallbackMode"`
	RpcFallbackCacheMode        int64  `json:"rpcFallbackCacheMode"`
	RpcRespFallbackClassName    string `json:"rpcRespFallbackClassName"`
	RpcFallbackExceptionMessage string `json:"rpcFallbackExceptionMessage"`
	RpcRespContentBody          string `json:"rpcRespContentBody"`
}
