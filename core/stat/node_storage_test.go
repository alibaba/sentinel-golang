package stat

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInboundNode(t *testing.T) {
	n := InboundNode()

	assert.NotNil(t, n)
	assert.Equal(t, base.TotalInBoundResourceName, n.resourceName)
	assert.Equal(t, base.ResTypeCommon, n.resourceType)
}

func TestResetResourceNodeMap(t *testing.T) {
	t.Run("OriginEmptyResNodeMap", func(t *testing.T) {
		ResetResourceNodeMap()
		assert.Equal(t, 0, len(resNodeMap))
	})

	t.Run("OrinfinNonEmptyResNodeMap", func(t *testing.T) {
		resNodeMap = map[string]*ResourceNode{
			"test1": nil,
		}
		assert.Equal(t, 1, len(resNodeMap))
		ResetResourceNodeMap()
		assert.Equal(t, 0, len(resNodeMap))
	})
}

func TestGetResourceNode(t *testing.T) {
	t.Run("EmptyKeyResNodeMap", func(t *testing.T) {
		assert.Nil(t, GetResourceNode("test1"))
	})

	t.Run("EmptyValueResNodeMap", func(t *testing.T) {
		defer func() { resNodeMap = make(ResourceNodeMap, 0) }()

		resNodeMap["test1"] = nil
		assert.Nil(t, GetResourceNode("test1"))
	})

	t.Run("NormalResNodeMap", func(t *testing.T) {
		defer func() { resNodeMap = make(ResourceNodeMap, 0) }()

		n := &ResourceNode{}
		resNodeMap["test1"] = n
		assert.Equal(t, n, GetResourceNode("test1"))
	})
}

func TestGetOrCreateResourceNode(t *testing.T) {
	n := NewResourceNode("test1", base.ResTypeCommon)

	t.Run("ExistKey", func(t *testing.T) {
		defer func() { resNodeMap = make(ResourceNodeMap, 0) }()

		resNodeMap["test1"] = n
		assert.Equal(t, n, GetOrCreateResourceNode("test1", base.ResTypeCommon))
	})

	t.Run("NonExistKey", func(t *testing.T) {
		defer func() { resNodeMap = make(ResourceNodeMap, 0) }()
		gn := GetOrCreateResourceNode("test1", base.ResTypeCommon)
		assert.Equal(t, 1, len(resNodeMap))
		assert.Equal(t, gn, resNodeMap["test1"])
	})
}

func TestResourceNodeList(t *testing.T) {
	t.Run("EmptyResNodeMap", func(t *testing.T) {
		list := ResourceNodeList()
		assert.Equal(t, 0, len(list))
	})

	t.Run("NormalResNodeMap", func(t *testing.T) {
		defer func() { resNodeMap = make(ResourceNodeMap, 0) }()

		n := NewResourceNode("test1", base.ResTypeCommon)
		resNodeMap["test1"] = n

		list := ResourceNodeList()
		assert.Equal(t, 1, len(list))
	})

}
