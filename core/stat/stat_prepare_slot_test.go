package stat

import (
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"

	"github.com/stretchr/testify/assert"
)

func TestStatNodePrepareSlot_Prepare(t *testing.T) {
	s := &StatNodePrepareSlot{}
	rw := base.NewResourceWrapper("test1", base.ResTypeCommon, base.Inbound)
	ctx := &base.EntryContext{
		Resource: rw,
	}

	s.Prepare(ctx)
	assert.Equal(t, resNodeMap["test1"], ctx.StatNode)
}
