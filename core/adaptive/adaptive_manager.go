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
type ControllerGenFunc func(r *Config) Controller

var (
	acMap             = make(map[string]Controller, 0)
	controllerGenFunc = make(map[AdaptiveType]ControllerGenFunc, 4)
	updateConfigMux   = new(sync.Mutex)
	acMux             = new(sync.RWMutex)
	currentConfigs    = make([]*Config, 0)
)

func init() {
	// Initialize the controller generator map for existing adaptive types.
	controllerGenFunc[Memory] = func(c *Config) Controller {
		return newMemoryAdaptiveController(c)
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

	acMux.RLock()
	m := make(map[string]Controller, len(validConfigs))
	for _, config := range validConfigs {
		c, b := acMap[config.AdaptiveConfigName]
		if b && c.BoundConfig().IsEqualsTo(config) {
			m[config.AdaptiveConfigName] = c
			continue
		}
		generator, supported := controllerGenFunc[config.AdaptiveType]
		if !supported {
			logging.Warn("[Adaptive onConfigUpdate] Ignoring the adaptive config due to unsupported adaptive type", "config", config)
			continue
		}
		m[config.AdaptiveConfigName] = generator(config)
	}

	acMux.RUnlock()

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
	if config.AdaptiveConfigName == "" {
		return errors.New("empty AdaptiveConfigName")
	}
	if config.AdaptiveType != Memory {
		return errors.New("invalid AdaptiveType")
	}
	if config.AdaptiveType == Memory {
		if config.LowRatio <= 0 {
			return errors.New("config.LowRatio <= 0")
		}
		if config.HighRatio <= 0 {
			return errors.New("config.HighRatio <= 0")
		}
		if config.HighRatio >= config.LowRatio {
			return errors.New("config.HighRatio >= config.LowRatio")
		}

		if config.LowWaterMark <= 0 {
			return errors.New("config.LowWaterMark <= 0")
		}
		if config.HighWaterMark <= 0 {
			return errors.New("config.HighWaterMark <= 0")
		}
		if int64(config.HighWaterMark) > int64(system_metric.TotalMemorySize) {
			return errors.New("config.HighWaterMark should not be greater than current system's total memory size")
		}
		if config.LowWaterMark >= config.HighWaterMark {
			// can not be equal to defeat from zero overflow
			return errors.New("config.LowWaterMark >= config.HighWaterMark")
		}
	}

	return nil
}
