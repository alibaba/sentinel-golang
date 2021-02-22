package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBlockError(t *testing.T) {
	type args struct {
		opts []BlockErrorOption
	}
	tests := []struct {
		name string
		args args
		want *BlockError
	}{
		{
			name: "normal",
			args: struct{ opts []BlockErrorOption }{opts: []BlockErrorOption{
				WithBlockType(BlockTypeFlow),
				WithBlockMsg("test"),
				WithRule(new(MockRule)),
				WithSnapshotValue("snapshot"),
			}},
			want: NewBlockError(WithBlockType(BlockTypeFlow),
				WithBlockMsg("test"),
				WithRule(new(MockRule)),
				WithSnapshotValue("snapshot")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBlockError(tt.args.opts...)
			assert.Equal(t, got.blockType, BlockTypeFlow)
			assert.Equal(t, got.blockMsg, "test")
			assert.Equal(t, got.rule, new(MockRule))
			assert.Equal(t, got.snapshotValue, "snapshot")
		})
	}
}

type MockRule struct {
	Id string `json:"id"`
}

func (m *MockRule) String() string {
	return "mock rule"
}

func (m *MockRule) ResourceName() string {
	return "mock resource"
}
