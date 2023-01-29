package base

import (
	"reflect"
	"testing"
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
			if got := NewBlockError(tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockError() = %v, want %v", got, tt.want)
			}
		})
	}
}
