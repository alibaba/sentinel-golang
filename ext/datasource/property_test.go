package datasource

import (
	"encoding/json"
	"errors"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func MockSystemRulesConvert(src []byte) interface{} {
	ret := make([]system.SystemRule, 0)
	_ = json.Unmarshal(src, &ret)
	return ret
}
func MockSystemRulesConvertReturnNil(src []byte) interface{} {
	return nil
}
func MockSystemRulesUpdateReturnNil(data interface{}) error {
	return nil
}
func MockSystemRulesUpdateReturnError(data interface{}) error {
	return errors.New("MockSystemRulesUpdateReturnError")
}

func TestNewSinglePropertyHandler(t *testing.T) {
	got := NewSinglePropertyHandler(MockSystemRulesConvert, MockSystemRulesUpdateReturnNil)
	assert.Truef(t, got.lastUpdateProperty == nil, "lastUpdatePropertyHash:%d, expect nil", got.lastUpdateProperty)
}

func TestSinglePropertyHandler_Handle(t *testing.T) {
	h1 := NewSinglePropertyHandler(MockSystemRulesConvertReturnNil, MockSystemRulesUpdateReturnNil)
	r1 := h1.Handle(nil)
	assert.True(t, r1 == nil, "Fail to execute Handle func.")

	h2 := NewSinglePropertyHandler(MockSystemRulesConvert, MockSystemRulesUpdateReturnError)
	src, err := ioutil.ReadFile("../../tests/testdata/extension/SystemRule.json")
	if err != nil {
		t.Errorf("Fail to get source file, err:%+v", err)
	}
	r2 := h2.Handle(src)
	assert.True(t, r2 != nil&&r2.Error()=="MockSystemRulesUpdateReturnError", "Fail to execute Handle func.")
}

func TestSinglePropertyHandler_isPropertyConsistent(t *testing.T) {
	h := NewSinglePropertyHandler(MockSystemRulesConvert, MockSystemRulesUpdateReturnNil)
	src, err := ioutil.ReadFile("../../tests/testdata/extension/SystemRule.json")
	if err != nil {
		t.Errorf("Fail to get source file, err:%+v", err)
	}
	ret1 := make([]system.SystemRule, 0)
	_ = json.Unmarshal(src, &ret1)
	isConsistent := h.isPropertyConsistent(ret1)
	assert.True(t, isConsistent == false, "Fail to execute isPropertyConsistent.")

	src2, err := ioutil.ReadFile("../../tests/testdata/extension/SystemRule2.json")
	if err != nil {
		t.Errorf("Fail to get source file, err:%+v", err)
	}
	ret2 := make([]system.SystemRule, 0)
	_ = json.Unmarshal(src2, &ret2)
	isConsistent = h.isPropertyConsistent(ret2)
	assert.True(t, isConsistent == true, "Fail to execute isPropertyConsistent.")

	src3, err := ioutil.ReadFile("../../tests/testdata/extension/SystemRule3.json")
	if err != nil {
		t.Errorf("Fail to get source file, err:%+v", err)
	}
	ret3 := make([]system.SystemRule, 0)
	_ = json.Unmarshal(src3, &ret3)
	isConsistent = h.isPropertyConsistent(ret3)
	assert.True(t, isConsistent == false, "Fail to execute isPropertyConsistent.")
}
