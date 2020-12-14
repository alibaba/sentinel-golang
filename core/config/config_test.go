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

package config

import (
	"os"
	"testing"
)

const testDataBaseDir = "../../tests/testdata/config/"

func TestLoadFromYamlFile(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestLoadFromYamlFile",
			args: args{
				filePath: testDataBaseDir + "sentinel.yml",
			},
			wantErr: false,
		},
		{
			name: "TestLoadFromYamlFile",
			args: args{
				filePath: testDataBaseDir + "sentinel.yml.1",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := loadGlobalConfigFromYamlFile(tt.args.filePath); (err != nil) != tt.wantErr {
				t.Errorf("loadGlobalConfigFromYamlFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOverrideFromSystemEnv(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "TestOverrideFromSystemEnv",
			wantErr: false,
		},
	}
	err := loadGlobalConfigFromYamlFile(testDataBaseDir + "sentinel.yml")
	if err != nil {
		t.Errorf("Fail to initialize data.")
	}
	_ = os.Setenv(AppNameEnvKey, "app-name")
	_ = os.Setenv(AppTypeEnvKey, "1")
	_ = os.Setenv(LogDirEnvKey, testDataBaseDir+"sentinel.yml.2")
	_ = os.Setenv(LogNamePidEnvKey, "true")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := overrideItemsFromSystemEnv(); (err != nil) != tt.wantErr {
				t.Errorf("overrideItemsFromSystemEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
