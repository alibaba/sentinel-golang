package util

import "time"

func GetTimeMilli() uint64 {
	return uint64(time.Now().UnixNano() / 1e6)
}

//noinspection GoUnusedExportedFunction
func GetTimeNano() uint64 {
	return uint64(time.Now().UnixNano())
}

//noinspection GoUnusedExportedFunction
func GetTimeSecond() uint64 {
	return uint64(time.Now().Unix())
}
