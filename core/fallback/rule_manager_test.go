package fallback

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadRule(t *testing.T) {
	_, err := LoadRules([]*Rule{
		{
			TargetResourceType: WebResourceType,
			TargetMap: map[string][]FunctionType{
				"/greet": {
					FlowType,
					Isolation,
				},
			},
			FallbackBehavior: []byte("{\"webFallbackMode\":0,\"webRespContentType\":1,\"webRespMessage\":\"{\\n  \\\"abc\\\": 123\\n}\",\"webRespStatusCode\":433}"),
		},
		{
			TargetResourceType: WebResourceType,
			TargetMap: map[string][]FunctionType{
				"/greet": {
					HotspotHttp,
				},
			},
			FallbackBehavior: []byte("{\"webFallbackMode\":0,\"webRespContentType\":1,\"webRespMessage\":\"{\\n  \\\"abc\\\": 123\\n}\",\"webRespStatusCode\":434}"),
		},
		{
			TargetResourceType: WebResourceType,
			TargetMap: map[string][]FunctionType{
				"/api/users/:id": {
					FlowType,
				},
			},
			FallbackBehavior: []byte("{\"webFallbackMode\":0,\"webRespContentType\":1,\"webRespMessage\":\"{\\n  \\\"abc\\\": 123\\n}\",\"webRespStatusCode\":400}"),
		},
	})
	assert.NoError(t, err)

	assert.Equal(t, 2, len(webRuleMap))
	assert.Equal(t, 0, len(rpcRuleMap))

	funcTypeMap := webRuleMap["/greet"]
	assert.Equal(t, 3, len(funcTypeMap))

	assert.Equal(t, funcTypeMap[FlowType].WebRespStatusCode, int64(433))
	assert.Equal(t, funcTypeMap[HotspotHttp].WebRespStatusCode, int64(434))

	funcTypeMap = webRuleMap["/api/users/:id"]
	assert.Equal(t, 1, len(funcTypeMap))

}
