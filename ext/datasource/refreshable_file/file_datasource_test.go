package refreshable_file

import (
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	TestFlowRules = `[
    {
        "id": 0,
        "resource": "abc0",
        "limitApp": "default",
        "grade": 1,
        "strategy": 0,
        "controlBehavior": 0,
        "refResource": "refDefault",
        "warmUpPeriodSec": 10,
        "maxQueueingTimeMs": 1000,
        "clusterMode": false,
        "clusterConfig": {
            "thresholdType": 0
        }
    },
    {
        "id": 1,
        "resource": "abc1",
        "limitApp": "default",
        "grade": 1,
        "strategy": 0,
        "controlBehavior": 0,
        "refResource": "refDefault",
        "warmUpPeriodSec": 10,
        "maxQueueingTimeMs": 1000,
        "clusterMode": false,
        "clusterConfig": {
            "thresholdType": 0
        }
    },
    {
        "id": 2,
        "resource": "abc2",
        "limitApp": "default",
        "grade": 1,
        "strategy": 0,
        "controlBehavior": 0,
        "refResource": "refDefault",
        "warmUpPeriodSec": 10,
        "maxQueueingTimeMs": 1000,
        "clusterMode": false,
        "clusterConfig": {
            "thresholdType": 0
        }
    }
]`
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
	TestFlowRulesFile   = "../../../tests/testdata/extension/refreshable_file/FlowRule.json"
	TestSystemRulesFile = "../../../tests/testdata/extension/refreshable_file/FlowRule.json"
)

func prepareFlowRulesTestFile() error {
	content := []byte(TestFlowRules)
	return ioutil.WriteFile(TestFlowRulesFile, content, os.ModePerm)
}

func deleteFlowRulesTestFile() error {
	return os.Remove(TestFlowRulesFile)
}

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
		assert.True(t, got == nil && err != nil && strings.Contains(err.Error(), "Fail to open the property file"))

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
	err := prepareSystemRulesTestFile()
	if err != nil {
		t.Errorf("Fail to prepare test file, err: %+v", err)
	}

	mh1 := &datasource.MockPropertyHandler{}
	mh1.On("Handle", tmock.Anything).Return(nil)
	mh1.On("isPropertyConsistent", tmock.Anything).Return(false)

	ds, _ := NewFileDataSource(TestSystemRulesFile, mh1)
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
	e := ds.watcher.Add(TestSystemRulesFile)
	assert.True(t, e != nil && strings.Contains(e.Error(), "closed"))

	err = deleteSystemRulesTestFile()
	if err != nil {
		t.Errorf("Fail to delete test file, err: %+v", err)
	}
}
