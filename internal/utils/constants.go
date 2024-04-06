package utils

import (
	"math"
	"time"
)

var (
	INFINITY = time.Time.Add(time.Now(), time.Duration(math.MaxInt64))
)

const (
	DUMP_FILE_NAME   = "dump.txt"
	CLEANUP_DURATION = time.Duration(1 * time.Minute)
)
