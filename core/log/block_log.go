package log

import "github.com/alibaba/sentinel-golang/core/statlogger"

//200M
const MaxBlockLogFileSize = 200 * 1024 * 1024

var (
	BlockLogger = statlogger.NewStatLogger("sentinel-block.log", 3, 1000, 6000, MaxBlockLogFileSize)
)

func statBlockedEntry(batchCount uint32, resourceName string, blockError string) {
	BlockLogger.Stat(batchCount, resourceName, blockError)
}
