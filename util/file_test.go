package util

import (
	"testing"
)

func TestFileExists(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantB   bool
		wantErr bool
	}{
		{
			"file.go", args{"file.go"}, true, false,
		},
		{
			"file.gox", args{"file.gox"}, false, false,
		},
		{
			"empty", args{""}, false, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotB, err := FileExists(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotB != tt.wantB {
				t.Errorf("FileExists() = %v, want %v", gotB, tt.wantB)
			}
		})
	}
}
