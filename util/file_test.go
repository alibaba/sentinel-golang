// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"os"
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

func TestFileRename(t *testing.T) {
	type args struct {
		newNameSuffix string
		before        func() string
		clean         func(tempName string)
	}
	oldName := "testfile"
	newName := ".renamed"
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "rename success",
			want: true,
			args: args{
				newNameSuffix: newName,
				before: func() string {
					oldFile, err := os.CreateTemp("", oldName)
					if err != nil {
						t.Fatalf("Failed to create temporary file: %s", err)
					}
					tempName := oldFile.Name()
					return tempName
				},
				clean: func(tempName string) {
					os.RemoveAll(tempName)
				},
			},
		},
		{
			name: "rename fail",
			want: false,
			args: args{
				newNameSuffix: newName,
				before: func() string {
					err := os.RemoveAll(oldName)
					if err != nil {
						t.Fatalf("Failed to remove temporary file: %s", err)
					}
					return ""
				},
				clean: func(tempName string) {

				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oldFileName string
			if tt.args.before != nil {
				oldFileName = tt.args.before()
			}
			defer tt.args.clean(oldFileName)
			if got := FileRename(oldFileName, oldFileName+tt.args.newNameSuffix); got != tt.want {
				t.Errorf("FileRename() = %v, want %v", got, tt.want)
			}
		})
	}
}
