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
			if err := loadFromYamlFile(tt.args.filePath); (err != nil) != tt.wantErr {
				t.Errorf("loadFromYamlFile() error = %v, wantErr %v", err, tt.wantErr)
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
	err := loadFromYamlFile(testDataBaseDir + "sentinel.yml")
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
