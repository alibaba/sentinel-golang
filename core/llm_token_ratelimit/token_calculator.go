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

var globalTokenCalculator = NewDefaultTokenCalculator()

type TokenCalculator interface {
	Calculate(ctx *Context, infos *UsedTokenInfos) int
}

type TokenCalculatorManager struct {
	calculators map[CountStrategy]TokenCalculator
}

func NewDefaultTokenCalculator() *TokenCalculatorManager {
	return &TokenCalculatorManager{
		calculators: map[CountStrategy]TokenCalculator{
			TotalTokens:  &TotalTokensCalculator{},
			InputTokens:  &InputTokensCalculator{},
			OutputTokens: &OutputTokensCalculator{},
		}}
}

func (m *TokenCalculatorManager) getCalculator(strategy CountStrategy) TokenCalculator {
	if m == nil || m.calculators == nil {
		return nil
	}
	calculator, exists := m.calculators[strategy]
	if !exists {
		return nil
	}
	return calculator
}

type InputTokensCalculator struct{}

func (c *InputTokensCalculator) Calculate(ctx *Context, infos *UsedTokenInfos) int {
	if c == nil || infos == nil {
		return 0
	}
	return infos.InputTokens
}

type OutputTokensCalculator struct{}

func (c *OutputTokensCalculator) Calculate(ctx *Context, infos *UsedTokenInfos) int {
	if c == nil || infos == nil {
		return 0
	}
	return infos.OutputTokens
}

type TotalTokensCalculator struct{}

func (c *TotalTokensCalculator) Calculate(ctx *Context, infos *UsedTokenInfos) int {
	if c == nil || infos == nil {
		return 0
	}
	return infos.TotalTokens
}
