package utils

import (
	"math"
	"time"
)

var (
	INFINITY = time.Time.Add(time.Now(), time.Duration(math.MaxInt64))
)

const (
	DUMP_FILE_NAME   = "snapshot.txt"
	CLEANUP_DURATION = time.Duration(1 * time.Minute)
	TRANSACTIONAL    = 0
	ACTIVE           = 1
	CONFIG_FILE      = "kv.conf"
	PUB_SUB          = 2
)
