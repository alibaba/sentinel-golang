package stat

import "testing"

func TestRemoveResourceNodes(t *testing.T) {
	type args struct {
		resources []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "empty",
			args: args{},
		},
		{
			name: "normal",
			args: args{
				resources: []string{"test"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RemoveResourceNodes(tt.args.resources)
		})
	}
}
