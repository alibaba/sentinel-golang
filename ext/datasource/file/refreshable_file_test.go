package file

import (
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
)

const (
	TestSystemRules = `[
    {
        "id": 0,
        "metricType": 0,
        "adaptiveStrategy": 0
    },
    {
        "id": 1,
        "metricType": 0,
        "adaptiveStrategy": 0
    },
    {
        "id": 2,
        "metricType": 0,
        "adaptiveStrategy": 0
    }
]`
)

var (
	TestSystemRulesDir  = "./"
	TestSystemRulesFile = TestSystemRulesDir + "SystemRules.json"
)

func prepareSystemRulesTestFile() error {
	content := []byte(TestSystemRules)
	return ioutil.WriteFile(TestSystemRulesFile, content, os.ModePerm)
}

func deleteSystemRulesTestFile() error {
	return os.Remove(TestSystemRulesFile)
}

func TestRefreshableFileDataSource_ReadSource(t *testing.T) {
	t.Run("RefreshableFileDataSource_ReadSource_Nil", func(t *testing.T) {
		err := prepareSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to prepare test file, err: %+v", err)
		}

		s := &RefreshableFileDataSource{
			sourceFilePath: TestSystemRulesFile + "NotExisted",
		}
		got, err := s.ReadSource()
		assert.True(t, got == nil && err != nil && strings.Contains(err.Error(), "RefreshableFileDataSource fail to open the property file"))

		err = deleteSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to delete test file, err: %+v", err)
		}
	})

	t.Run("RefreshableFileDataSource_ReadSource_Normal", func(t *testing.T) {
		err := prepareSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to prepare test file, err: %+v", err)
		}

		s := &RefreshableFileDataSource{
			sourceFilePath: TestSystemRulesFile,
		}
		got, err := s.ReadSource()
		if err != nil {
			t.Errorf("Fail to execute ReadSource, err: %+v", err)
		}
		assert.True(t, reflect.DeepEqual(got, []byte(TestSystemRules)))

		err = deleteSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to delete test file, err: %+v", err)
		}
	})
}

func TestRefreshableFileDataSource_doReadAndUpdate(t *testing.T) {
	t.Run("TestRefreshableFileDataSource_doReadAndUpdate_normal", func(t *testing.T) {
		err := prepareSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to prepare test file, err: %+v", err)
		}

		s := &RefreshableFileDataSource{
			sourceFilePath: TestSystemRulesFile,
			closeChan:      make(chan struct{}),
		}
		mh1 := &datasource.MockPropertyHandler{}
		mh1.On("Handle", tmock.Anything).Return(nil)
		mh1.On("isPropertyConsistent", tmock.Anything).Return(false)
		s.AddPropertyHandler(mh1)

		err = s.doReadAndUpdate()
		assert.True(t, err == nil, "Fail to doReadAndUpdate.")

		err = deleteSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to delete test file, err: %+v", err)
		}
	})

	t.Run("TestRefreshableFileDataSource_doReadAndUpdate_Handler_err", func(t *testing.T) {
		err := prepareSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to prepare test file, err: %+v", err)
		}

		s := &RefreshableFileDataSource{
			sourceFilePath: TestSystemRulesFile,
			closeChan:      make(chan struct{}),
		}
		mh1 := &datasource.MockPropertyHandler{}
		hErr := errors.New("Handle error")
		mh1.On("Handle", tmock.Anything).Return(hErr)
		mh1.On("isPropertyConsistent", tmock.Anything).Return(false)
		s.AddPropertyHandler(mh1)

		err = s.doReadAndUpdate()
		assert.True(t, err != nil && strings.Contains(err.Error(), hErr.Error()), "Fail to doReadAndUpdate.")

		err = deleteSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to delete test file, err: %+v", err)
		}
	})

	t.Run("TestRefreshableFileDataSource_doReadAndUpdate_Multi_Handler_err", func(t *testing.T) {
		err := prepareSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to prepare test file, err: %+v", err)
		}

		s := &RefreshableFileDataSource{
			sourceFilePath: TestSystemRulesFile,
			closeChan:      make(chan struct{}),
		}
		mh1 := &datasource.MockPropertyHandler{}
		hErr := errors.New("Handle error")
		mh1.On("Handle", tmock.Anything).Return(hErr)
		mh1.On("isPropertyConsistent", tmock.Anything).Return(false)
		mh2 := &datasource.MockPropertyHandler{}
		mh2.On("Handle", tmock.Anything).Return(nil)
		mh2.On("isPropertyConsistent", tmock.Anything).Return(false)

		s.AddPropertyHandler(mh1)
		s.AddPropertyHandler(mh2)

		err = s.doReadAndUpdate()

		mh1.AssertNumberOfCalls(t, "Handle", 1)
		mh2.AssertNumberOfCalls(t, "Handle", 1)

		assert.True(t, err != nil && strings.Contains(err.Error(), hErr.Error()), "Fail to doReadAndUpdate.")

		err = deleteSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to delete test file, err: %+v", err)
		}
	})
}

func TestRefreshableFileDataSource_Close(t *testing.T) {
	t.Run("TestRefreshableFileDataSource_Close", func(t *testing.T) {
		err := prepareSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to prepare test file, err: %+v", err)
		}

		s := &RefreshableFileDataSource{
			sourceFilePath: TestSystemRulesFile,
			closeChan:      make(chan struct{}),
		}
		mh1 := &datasource.MockPropertyHandler{}
		mh1.On("Handle", tmock.Anything).Return(nil)
		mh1.On("isPropertyConsistent", tmock.Anything).Return(false)
		s.AddPropertyHandler(mh1)

		err = s.Initialize()
		if err != nil {
			t.Errorf("Fail to Initialize datasource, err: %+v", err)
		}

		time.Sleep(1 * time.Second)
		s.Close()
		time.Sleep(1 * time.Second)
		e := s.watcher.Add(TestSystemRulesFile)
		assert.True(t, e != nil && strings.Contains(e.Error(), "closed"))

		err = deleteSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to delete test file, err: %+v", err)
		}
	})
}

func TestNewFileDataSource_ALL_For_SystemRule(t *testing.T) {
	t.Run("TestNewFileDataSource_ALL_For_SystemRule_Write_Event", func(t *testing.T) {
		err := prepareSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to prepare test file, err: %+v", err)
		}

		mh1 := &datasource.MockPropertyHandler{}
		mh1.On("Handle", tmock.Anything).Return(nil)
		mh1.On("isPropertyConsistent", tmock.Anything).Return(false)

		ds := NewFileDataSource(TestSystemRulesFile, mh1)
		err = ds.Initialize()
		if err != nil {
			t.Errorf("Fail to initialize the file data source, err: %+v", err)
		}
		mh1.AssertNumberOfCalls(t, "Handle", 1)

		f, err := os.OpenFile(ds.sourceFilePath, os.O_RDWR|os.O_APPEND|os.O_SYNC, os.ModePerm)
		if err != nil {
			t.Errorf("Fail to open the property file, err: %+v.", err)
		}
		defer f.Close()

		f.WriteString("\n" + TestSystemRules)
		f.Sync()
		time.Sleep(3 * time.Second)
		mh1.AssertNumberOfCalls(t, "Handle", 2)

		ds.Close()
		f.Close()
		time.Sleep(1 * time.Second)
		e := ds.watcher.Add(TestSystemRulesFile)
		assert.True(t, e != nil && strings.Contains(e.Error(), "closed"))

		err = deleteSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to delete test file, err: %+v", err)
		}
	})

	t.Run("TestNewFileDataSource_ALL_For_SystemRule_Remove_Event", func(t *testing.T) {
		err := prepareSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to prepare test file, err: %+v", err)
		}

		mh1 := &datasource.MockPropertyHandler{}
		mh1.On("Handle", tmock.Anything).Return(nil)
		mh1.On("isPropertyConsistent", tmock.Anything).Return(false)

		ds := NewFileDataSource(TestSystemRulesFile, mh1)
		err = ds.Initialize()
		if err != nil {
			t.Errorf("Fail to initialize the file data source, err: %+v", err)
		}
		mh1.AssertNumberOfCalls(t, "Handle", 1)

		err = deleteSystemRulesTestFile()
		if err != nil {
			t.Errorf("Fail to delete test file, err: %+v", err)
		}

		time.Sleep(3 * time.Second)
		mh1.AssertNumberOfCalls(t, "Handle", 2)

		ds.Close()
		time.Sleep(1 * time.Second)
		e := ds.watcher.Add(TestSystemRulesFile)
		assert.True(t, e != nil && strings.Contains(e.Error(), "closed"))
	})

}
