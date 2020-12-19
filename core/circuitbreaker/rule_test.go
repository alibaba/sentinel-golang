package circuitbreaker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuleIsStatReusable(t *testing.T) {
	cases := []struct {
		rule1          *Rule
		rule2          *Rule
		expectedResult bool
	}{
		// nil
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2:          nil,
			expectedResult: false,
		},

		// different Resource
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2: &Rule{
				Resource:                     "def",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			expectedResult: false,
		},

		// different Strategy
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorRatio,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    0.5,
			},
			expectedResult: false,
		},

		// different StatIntervalMs
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               5000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			expectedResult: false,
		},

		// different StatSlidingWindowBucketCount
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 5,
				Threshold:                    1.0,
			},
			expectedResult: false,
		},

		// different RetryTimeoutMs
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               5000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			expectedResult: true,
		},

		// different MinRequestAmount
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             20,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			expectedResult: true,
		},

		// different Threshold
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             20,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    2.0,
			},
			expectedResult: true,
		},
	}

	for i, c := range cases {
		result := c.rule1.isStatReusable(c.rule2)
		assert.Equal(t, c.expectedResult, result, fmt.Sprintf("case %d got unexpected result", i))
	}
}

func TestRuleIsEqualsToBase(t *testing.T) {
	cases := []struct {
		rule1          *Rule
		rule2          *Rule
		expectedResult bool
	}{
		// nil
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2:          nil,
			expectedResult: false,
		},

		// different Resource
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2: &Rule{
				Resource:                     "def",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			expectedResult: false,
		},

		// different Strategy
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorRatio,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    0.5,
			},
			expectedResult: false,
		},

		// different StatIntervalMs
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               5000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			expectedResult: false,
		},

		// different StatSlidingWindowBucketCount
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 5,
				Threshold:                    1.0,
			},
			expectedResult: false,
		},

		// different RetryTimeoutMs
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               5000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			expectedResult: false,
		},

		// different MinRequestAmount
		{
			rule1: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			rule2: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             20,
				StatIntervalMs:               10000,
				StatSlidingWindowBucketCount: 2,
				Threshold:                    1.0,
			},
			expectedResult: false,
		},
	}

	for i, c := range cases {
		result := c.rule1.isEqualsToBase(c.rule2)
		assert.Equal(t, c.expectedResult, result, fmt.Sprintf("case %d got unexpected result", i))
	}
}

func TestRuleGetValidSlidingWindowBucketCount(t *testing.T) {
	cases := []struct {
		rule                *Rule
		expectedBucketCount uint32
	}{
		{
			rule: &Rule{
				Resource:         "abc",
				Strategy:         ErrorCount,
				RetryTimeoutMs:   3000,
				MinRequestAmount: 10,
				StatIntervalMs:   10000,
				Threshold:        1.0,
			},
			expectedBucketCount: 1,
		},
		{
			rule: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               1000,
				StatSlidingWindowBucketCount: 1,
				Threshold:                    1.0,
			},
			expectedBucketCount: 1,
		},
		{
			rule: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               1000,
				StatSlidingWindowBucketCount: 10,
				Threshold:                    1.0,
			},
			expectedBucketCount: 10,
		},
		{
			rule: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               1000,
				StatSlidingWindowBucketCount: 30,
				Threshold:                    1.0,
			},
			expectedBucketCount: 1,
		},
		{
			rule: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               100,
				StatSlidingWindowBucketCount: 100,
				Threshold:                    1.0,
			},
			expectedBucketCount: 100,
		},
		{
			rule: &Rule{
				Resource:                     "abc",
				Strategy:                     ErrorCount,
				RetryTimeoutMs:               3000,
				MinRequestAmount:             10,
				StatIntervalMs:               100,
				StatSlidingWindowBucketCount: 200,
				Threshold:                    1.0,
			},
			expectedBucketCount: 1,
		},
	}

	for i, c := range cases {
		bucketCount := getStatSlidingWindowBucketCount(c.rule)
		assert.Equal(t, c.expectedBucketCount, bucketCount, fmt.Sprintf("case %d got unexpected result", i))
	}
}
