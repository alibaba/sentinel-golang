package base

import (
	"errors"
	"strconv"
	"strings"

	"github.com/alibaba/sentinel-golang/util"
)

const metricPartSeparator = "|"

// MetricItem represents the data of metric log per line.
type MetricItem struct {
	Resource       string
	Classification int32
	Timestamp      uint64

	PassQps         uint64
	BlockQps        uint64
	CompleteQps     uint64
	ErrorQps        uint64
	AvgRt           uint64
	OccupiedPassQps uint64
	Concurrency     uint32
}

type MetricItemRetriever interface {
	MetricsOnCondition(predicate TimePredicate) []*MetricItem
}

func (m *MetricItem) ToFatString() (string, error) {
	b := strings.Builder{}
	b.Grow(128)

	timeStr := util.FormatTimeMillis(m.Timestamp)
	// All "|" in the resource name will be replaced with "_"
	finalName := strings.ReplaceAll(m.Resource, "|", "_")

	b.WriteString(strconv.FormatUint(m.Timestamp, 10))
	b.WriteByte('|')
	b.WriteString(timeStr)
	b.WriteByte('|')
	b.WriteString(finalName)
	b.WriteByte('|')
	b.WriteString(strconv.FormatUint(m.PassQps, 10))
	b.WriteByte('|')
	b.WriteString(strconv.FormatUint(m.BlockQps, 10))
	b.WriteByte('|')
	b.WriteString(strconv.FormatUint(m.CompleteQps, 10))
	b.WriteByte('|')
	b.WriteString(strconv.FormatUint(m.ErrorQps, 10))
	b.WriteByte('|')
	b.WriteString(strconv.FormatUint(m.AvgRt, 10))
	b.WriteByte('|')
	b.WriteString(strconv.FormatUint(m.OccupiedPassQps, 10))
	b.WriteByte('|')
	b.WriteString(strconv.FormatUint(uint64(m.Concurrency), 10))
	b.WriteByte('|')
	b.WriteString(strconv.FormatInt(int64(m.Classification), 10))

	return b.String(), nil
}

func (m *MetricItem) ToThinString() (string, error) {
	b := strings.Builder{}
	b.Grow(128)

	// All "|" in the resource name will be replaced with "_"
	finalName := strings.ReplaceAll(m.Resource, "|", "_")

	b.WriteString(strconv.FormatUint(m.Timestamp, 10))
	b.WriteByte('|')
	b.WriteString(finalName)
	b.WriteByte('|')
	b.WriteString(strconv.FormatUint(m.PassQps, 10))
	b.WriteByte('|')
	b.WriteString(strconv.FormatUint(m.BlockQps, 10))
	b.WriteByte('|')
	b.WriteString(strconv.FormatUint(m.CompleteQps, 10))
	b.WriteByte('|')
	b.WriteString(strconv.FormatUint(m.ErrorQps, 10))
	b.WriteByte('|')
	b.WriteString(strconv.FormatUint(m.AvgRt, 10))
	b.WriteByte('|')
	b.WriteString(strconv.FormatUint(m.OccupiedPassQps, 10))
	b.WriteByte('|')
	b.WriteString(strconv.FormatUint(uint64(m.Concurrency), 10))
	b.WriteByte('|')
	b.WriteString(strconv.FormatInt(int64(m.Classification), 10))

	return b.String(), nil
}

func MetricItemFromFatString(line string) (*MetricItem, error) {
	if len(line) == 0 {
		return nil, errors.New("invalid metric line: empty string")
	}
	item := &MetricItem{}
	arr := strings.Split(line, metricPartSeparator)
	if len(arr) < 8 {
		return nil, errors.New("invalid metric line: invalid format")
	}
	ts, err := strconv.ParseUint(arr[0], 10, 64)
	if err != nil {
		return nil, err
	}
	item.Timestamp = ts
	item.Resource = arr[2]
	p, err := strconv.ParseUint(arr[3], 10, 64)
	if err != nil {
		return nil, err
	}
	item.PassQps = p
	b, err := strconv.ParseUint(arr[4], 10, 64)
	if err != nil {
		return nil, err
	}
	item.BlockQps = b
	c, err := strconv.ParseUint(arr[5], 10, 64)
	if err != nil {
		return nil, err
	}
	item.CompleteQps = c
	e, err := strconv.ParseUint(arr[6], 10, 64)
	if err != nil {
		return nil, err
	}
	item.ErrorQps = e
	rt, err := strconv.ParseUint(arr[7], 10, 64)
	if err != nil {
		return nil, err
	}
	item.AvgRt = rt

	if len(arr) >= 9 {
		oc, err := strconv.ParseUint(arr[8], 10, 64)
		if err != nil {
			return nil, err
		}
		item.OccupiedPassQps = oc
	}
	if len(arr) >= 10 {
		concurrency, err := strconv.ParseUint(arr[9], 10, 32)
		if err != nil {
			return nil, err
		}
		item.Concurrency = uint32(concurrency)
	}
	if len(arr) >= 11 {
		cl, err := strconv.ParseInt(arr[10], 10, 32)
		if err != nil {
			return nil, err
		}
		item.Classification = int32(cl)
	}
	return item, nil
}
