package system

import (
	"encoding/json"
	"errors"
	"fmt"
)

func SystemRulesConvert(src []byte) interface{} {
	if src == nil{
		return nil
	}
	rule := make([]*SystemRule, 0)
	err := json.Unmarshal(src, &rule)
	if err != nil{
		logger.Errorf("Parsing data failed:%v", err)
		return nil
	}
	return rule
}

func SystemRulesUpdate(data interface{}) error {
	if data == nil{
		_, err := LoadRules(nil)
		return err
	}
	val, ok := data.([]*SystemRule)
	if !ok{
		return errors.New(fmt.Sprintf("Invalid parameters: data:%+v", data))
	}
	_, err := LoadRules(val)
	return err
}
