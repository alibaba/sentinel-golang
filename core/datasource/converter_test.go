package datasource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConverter(t *testing.T) {
	type FlowRule struct {
		Resource string
	}

	var flowRuleStr = `
{
	"resource": "demo"
}
`

	{
		/*value*/
		converter := JsonParseArray(FlowRule{})
		val, err := converter.Convert([]byte(flowRuleStr))
		assert.Nil(t, err)
		if flowRule, ok := val.(*FlowRule); !ok {
			assert.Fail(t, "covert failed")
		} else {
			assert.NotNil(t, flowRule)
			assert.Equal(t, flowRule.Resource, "demo")
		}
	}

	{
		/*pointer*/
		converter := JsonParseArray(&FlowRule{})
		val, err := converter.Convert([]byte(flowRuleStr))
		assert.Nil(t, err)
		if flowRule, ok := val.(*FlowRule); !ok {
			assert.Fail(t, "covert failed")
		} else {
			assert.NotNil(t, flowRule)
			assert.Equal(t, flowRule.Resource, "demo")
		}
	}
}
