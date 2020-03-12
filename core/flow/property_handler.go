package flow

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

func FlowRulesConvert(src []byte) (interface{}, error) {
	if src == nil{
		return nil, nil
	}
	rule := make([]*FlowRule,0)
	err := json.Unmarshal(src, &rule)
	if err != nil{
		return nil, errors.Errorf("Fail to unmarshal source:%+v to []flow.FlowRule, err:%+v", src, err)
	}
	return rule, nil
}

func FlowRulesUpdate(data interface{}) error {
	if data == nil{
		_, err := LoadRules(nil)
		return err
	}
	val, ok := data.([]*FlowRule)
	if !ok{
		return errors.New(fmt.Sprintf("Invalid parameters: data:%+v", data))
	}
	_, err := LoadRules(val)
	return err
}