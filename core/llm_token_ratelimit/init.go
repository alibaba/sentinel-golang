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

package llm_token_ratelimit

import (
	"fmt"

	"github.com/alibaba/sentinel-golang/logging"
)

func Init(cfg *Config) error {
	if globalConfig == nil {
		return fmt.Errorf("global config is nil")
	}
	if globalRedisClient == nil {
		return fmt.Errorf("global redis client is nil")
	}
	if globalRuleMatcher == nil {
		return fmt.Errorf("global rule matcher is nil")
	}
	if globalTokenCalculator == nil {
		return fmt.Errorf("global token calculator is nil")
	}

	if cfg == nil {
		cfg = NewDefaultConfig()
		if cfg == nil {
			return fmt.Errorf("new default config failed")
		}
	}

	if !cfg.Enabled {
		logging.Warn("[LLMTokenRateLimit] llm token rate limit is disabled, please enable it in config if needed")
		return nil
	}

	cfg.setDefaultConfigOptions()

	if err := globalConfig.SetConfig(cfg); err != nil {
		return err
	}
	if err := globalRedisClient.Init(cfg.Redis); err != nil {
		return err
	}
	return nil
}
