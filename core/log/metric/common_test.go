package metric

import (
	"github.com/stretchr/testify/assert"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestFormMetricFileName(t *testing.T) {
	appName1 := "foo-test"
	appName2 := "foo.test"
	mf1 := FormMetricFileName(appName1, false)
	mf2 := FormMetricFileName(appName2, false)
	assert.Equal(t, "foo-test-metrics.log", mf1)
	assert.Equal(t, mf1, mf2)
	mf1Pid := FormMetricFileName(appName2, true)
	if !strings.HasSuffix(mf1Pid, strconv.Itoa(os.Getpid())) {
		t.Fatalf("Metric filename <%s> should end with the process id", mf1Pid)
	}
}

func TestFilenameComparatorNoPid(t *testing.T) {
	arr := []string{
		"metrics.log.2018-03-06",
		"metrics.log.2018-03-07",
		"metrics.log.2018-03-07.51",
		"metrics.log.2018-03-07.10",
		"metrics.log.2018-03-06.100",
	}
	expected := []string{
		"metrics.log.2018-03-06",
		"metrics.log.2018-03-06.100",
		"metrics.log.2018-03-07",
		"metrics.log.2018-03-07.10",
		"metrics.log.2018-03-07.51",
	}

	sort.Slice(arr, filenameComparator(arr))
	assert.Equal(t, expected, arr)
}
