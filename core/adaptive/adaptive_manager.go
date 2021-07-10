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

package adaptive

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/alibaba/sentinel-golang/core/system_metric"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/pkg/errors"
)

// ControllerGenFunc represents the Controller generator function of a specific adaptive type.
type ControllerGenFunc func(r *Config) (Controller, error)

var (
	acMap             = make(map[string]Controller, 0)
	controllerGenFunc = make(map[MetricType]ControllerGenFunc, 4)
	updateConfigMux   = new(sync.Mutex)
	acMux             = new(sync.RWMutex)
	currentConfigs    = make([]*Config, 0)
)

func init() {
	// Initialize the controller generator map for existing adaptive types.
	controllerGenFunc[Memory] = func(c *Config) (Controller, error) {
		if c.CalculateStrategy == Linear {
			return newMemoryLinearAdaptiveController(c), nil
		}
		return nil, errors.Errorf("invalid CalculateStrategy:%s in adaptive.controllerGenFunc[Memory]", c.CalculateStrategy)
	}
}

//GetAdaptiveController gets AdaptiveController by adaptive name.
func GetAdaptiveController(adaptiveName string) Controller {
	acMux.RLock()
	defer acMux.RUnlock()
	c, _ := acMap[adaptiveName]
	return c
}

// LoadAdaptiveConfigs replaces all old adaptive configs with the given configs.
// Return value:
//   bool: indicates whether the internal map has been changed;
//   error: indicates whether occurs the error.
func LoadAdaptiveConfigs(configs []*Config) (bool, error) {
	updateConfigMux.Lock()
	defer updateConfigMux.Unlock()
	isEqual := reflect.DeepEqual(currentConfigs, configs)
	if isEqual {
		logging.Info("[Adaptive] Load adaptive config is the same with current configs, so ignore load operation.")
		return false, nil
	}

	err := onConfigUpdate(configs)
	return true, err
}

func onConfigUpdate(rawConfigs []*Config) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%+v", r)
			}
		}
	}()

	validConfigs := make([]*Config, 0, len(rawConfigs))
	for _, config := range rawConfigs {
		if err := IsValidConfig(config); err != nil {
			logging.Warn("[Adaptive onConfigUpdate] Ignoring invalid adaptive config", "config", config, "reason", err.Error())
			continue
		}
		validConfigs = append(validConfigs, config)
	}

	start := util.CurrentTimeNano()

	m := make(map[string]Controller, len(validConfigs))
	for _, config := range validConfigs {
		generator, supported := controllerGenFunc[config.MetricType]
		if !supported {
			logging.Warn("[Adaptive onConfigUpdate] Ignoring the adaptive config due to unsupported adaptive type", "config", config)
			continue
		}

		controller, err := generator(config)
		if controller == nil || err != nil {
			logging.Error(err, "Ignoring the rule due to bad generated controller in adaptive.onConfigUpdate()", "config", config)
			continue
		}
		m[config.ConfigName] = controller
	}

	acMux.Lock()
	acMap = m
	acMux.Unlock()

	currentConfigs = rawConfigs
	logging.Debug("[Adaptive onConfigUpdate] Time statistic(ns) for updating adaptive configs", "timeCost", util.CurrentTimeNano()-start)

	if len(validConfigs) == 0 {
		logging.Info("[AdaptiveManager] Adaptive configs were cleared")
	} else {
		logging.Info("[AdaptiveManager] Adaptive configs configs loaded", "configs", validConfigs)
	}
	return nil
}

// IsValidConfig checks whether the given config is valid.
func IsValidConfig(config *Config) error {
	if config == nil {
		return errors.New("nil Config")
	}
	if config.ConfigName == "" {
		return errors.New("empty ConfigName")
	}
	if config.MetricType != Memory {
		return errors.New("invalid Type")
	}
	if config.CalculateStrategy != Linear {
		return errors.New("invalid CalculateStrategy")
	}
	if config.CalculateStrategy == Linear {
		if config.LinearStrategyParameters == nil {
			return errors.New("empty LinearStrategyParameters")
		}
		if config.LinearStrategyParameters.LowRatio <= 0 {
			return errors.New("config.LinearStrategyParameters.LowRatio <= 0")
		}
		if config.LinearStrategyParameters.HighRatio <= 0 {
			return errors.New("config.LinearStrategyParameters.HighRatio <= 0")
		}
		if config.LinearStrategyParameters.HighRatio >= config.LinearStrategyParameters.LowRatio {
			return errors.New("config.LinearStrategyParameters.HighRatio >= config.LinearStrategyParameters.LowRatio")
		}

		if config.LinearStrategyParameters.LowWaterMark <= 0 {
			return errors.New("config.LinearStrategyParameters.LowWaterMark <= 0")
		}
		if config.LinearStrategyParameters.HighWaterMark <= 0 {
			return errors.New("config.LinearStrategyParameters.HighWaterMark <= 0")
		}
		if int64(config.LinearStrategyParameters.HighWaterMark) > int64(system_metric.TotalMemorySize) {
			return errors.New("config.LinearStrategyParameters.HighWaterMark should not be greater than current system's total memory size")
		}
		if config.LinearStrategyParameters.LowWaterMark >= config.LinearStrategyParameters.HighWaterMark {
			// can not be equal to defeat from zero overflow
			return errors.New("config.LinearStrategyParameters.LowWaterMark >= config.LinearStrategyParameters.HighWaterMark")
		}
	}
	return nil
}
