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

package datasource

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func MockSystemRulesConverter(src []byte) (interface{}, error) {
	ret := make([]system.Rule, 0)
	_ = json.Unmarshal(src, &ret)
	return ret, nil
}
func MockSystemRulesConverterReturnNil(src []byte) (interface{}, error) {
	return nil, nil
}
func MockSystemRulesUpdaterReturnNil(data interface{}) error {
	return nil
}
func MockSystemRulesUpdaterReturnError(data interface{}) error {
	return errors.New("MockSystemRulesUpdaterReturnError")
}

func TestNewSinglePropertyHandler(t *testing.T) {
	got := NewDefaultPropertyHandler(MockSystemRulesConverter, MockSystemRulesUpdaterReturnNil)
	assert.Truef(t, got.lastUpdateProperty == nil, "lastUpdateProperty:%d, expect nil", got.lastUpdateProperty)
}

func TestSinglePropertyHandler_Handle(t *testing.T) {
	h1 := NewDefaultPropertyHandler(MockSystemRulesConverterReturnNil, MockSystemRulesUpdaterReturnNil)
	r1 := h1.Handle(nil)
	assert.True(t, r1 == nil, "Fail to execute Handle func.")

	h2 := NewDefaultPropertyHandler(MockSystemRulesConverter, MockSystemRulesUpdaterReturnError)
	src, err := ioutil.ReadFile("../../tests/testdata/extension/SystemRule.json")
	if err != nil {
		t.Errorf("Fail to get source file, err:%+v", err)
	}
	r2 := h2.Handle(src)
	assert.True(t, r2 != nil && r2.Error() == "MockSystemRulesUpdaterReturnError", "Fail to execute Handle func.")
}

func TestSinglePropertyHandler_isPropertyConsistent(t *testing.T) {
	h := NewDefaultPropertyHandler(MockSystemRulesConverter, MockSystemRulesUpdaterReturnNil)
	src, err := ioutil.ReadFile("../../tests/testdata/extension/SystemRule.json")
	if err != nil {
		t.Errorf("Fail to get source file, err:%+v", err)
	}
	ret1 := make([]system.Rule, 0)
	_ = json.Unmarshal(src, &ret1)
	isConsistent := h.isPropertyConsistent(ret1)
	assert.True(t, isConsistent == false, "Fail to execute isPropertyConsistent.")

	src2, err := ioutil.ReadFile("../../tests/testdata/extension/SystemRule2.json")
	if err != nil {
		t.Errorf("Fail to get source file, err:%+v", err)
	}
	ret2 := make([]system.Rule, 0)
	_ = json.Unmarshal(src2, &ret2)
	isConsistent = h.isPropertyConsistent(ret2)
	assert.True(t, isConsistent == true, "Fail to execute isPropertyConsistent.")

	src3, err := ioutil.ReadFile("../../tests/testdata/extension/SystemRule3.json")
	if err != nil {
		t.Errorf("Fail to get source file, err:%+v", err)
	}
	ret3 := make([]system.Rule, 0)
	_ = json.Unmarshal(src3, &ret3)
	isConsistent = h.isPropertyConsistent(ret3)
	assert.True(t, isConsistent == false, "Fail to execute isPropertyConsistent.")
}
