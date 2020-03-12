package system

import (
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
)

func SystemRulesConvert(src []byte) (interface{}, error) {
	if src == nil{
		return nil, nil
	}
	rule := make([]*SystemRule, 0)
	err := json.Unmarshal(src, &rule)
	if err != nil{
		logger.Errorf("Parsing data failed:%v", err)
		return nil, errors.Errorf("Fail to unmarshal source:%+v to []system.SystemRule, err:%+v", src, err)
	}
	return rule, nil
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
