package flow

import (
	"encoding/json"
	"errors"
	"fmt"
)

func FlowRulesConvert(src []byte) interface{} {
	if src == nil{
		return nil
	}
	rule := make([]*FlowRule,0)
	err := json.Unmarshal(src, &rule)
	if err != nil{
		logger.Errorf("Parsing data failed:%v", err)
		return nil
	}
	return rule
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