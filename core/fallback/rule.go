package fallback

type TargetResourceType int64

const (
	WebResourceType TargetResourceType = 1
	RpcResourceType TargetResourceType = 2
)

type FunctionType int64

const (
	FlowType    FunctionType = 1
	Isolation   FunctionType = 6
	HotspotRpc  FunctionType = 4
	HotspotHttp FunctionType = 11
)

type Rule struct {
	TargetResourceType TargetResourceType        `json:"targetResourceType"`
	TargetMap          map[string][]FunctionType `json:"targetMap"`
	FallbackBehavior   interface{}               `json:"fallbackBehavior"`
}

type WebBlockFallbackBehavior struct {
	WebFallbackMode    int64  `json:"webFallbackMode"` // 0: return, 1: redirect
	WebRespStatusCode  int64  `json:"webRespStatusCode"`
	WebRespMessage     string `json:"webRespMessage"`
	WebRespContentType int64  `json:"webRespContentType"` // 0: test, 1: json
	WebRedirectUrl     string `json:"webRedirectUrl"`
}

type RpcBlockFallbackBehavior struct {
	RpcFallbackMode             int64  `json:"rpcFallbackMode"`
	RpcFallbackCacheMode        int64  `json:"rpcFallbackCacheMode"`
	RpcRespFallbackClassName    string `json:"rpcRespFallbackClassName"`
	RpcFallbackExceptionMessage string `json:"rpcFallbackExceptionMessage"`
	RpcRespContentBody          string `json:"rpcRespContentBody"`
}
