package file_test

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/extension/datasource/file"
	"gotest.tools/assert"
)

var testingFilePath = os.TempDir()+"file_datasource.json"
var testingExpect Data
var testingDataSource base.DataSource

type Data struct {
	A int `json:"a" toml:"a" yaml:"a"`
	B string `json:"b" toml:"b" yaml:"b"`
}

func TestMain(m *testing.M) {
	testingDataSource = file.New(testingFilePath)
	m.Run()
	testingDataSource.Close()
}

func initRegister(t *testing.T) {
	base.RegisterPropertyConsumer(func(decoder base.PropertyDecoder) error {
		var data Data
		err :=  decoder.Decode(&data)
		if err != nil && err != io.EOF {
			t.Fatalf("err: %+v\n", err.Error())
		}
		assert.Equal(t, data, testingExpect)
		return err
	}, func() error {
		return nil
	})
}

func updateFile(t *testing.T, raw []byte, expect Data) {
	testingExpect = expect
	assert.NilError(t, ioutil.WriteFile(testingFilePath, raw, os.ModePerm))
}

func deleteFile(t *testing.T) {
	testingExpect = Data{0,""}
	assert.NilError(t, os.Remove(testingFilePath))
}

func TestFileDataSource(t *testing.T) {
	initRegister(t)

	// open file
	updateFile(t, []byte(`{"A":1, "B":"2"}`), Data{A:1,B:"2"})
	assert.NilError(t, testingDataSource.ReadConfig())

	// watch file change
	updateFile(t, []byte(`{"A":3, "B":"4"}`), Data{A:3,B:"4"})
	time.Sleep(time.Second)

	// delete file
	deleteFile(t)
	time.Sleep(time.Second)
}
