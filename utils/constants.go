package utils

import (
	"math"
	"time"
)

var (
	INFINITY           = time.Time.Add(time.Now(), time.Duration(math.MaxInt64))
	MASTER_CONFIG_FILE = CONF_DIRECTORY + "kv.conf"
	SLAVE_1_CONFIG     = CONF_DIRECTORY + "slave1.conf"
	SLAVE_2_CONFIG     = CONF_DIRECTORY + "slave2.conf"
	SLAVE_3_CONFIG     = CONF_DIRECTORY + "slave3.conf"
	SNAPSHOT_FILE      = SNAPSHOT_DIRECTORY + "snapshot.txt"
	SHARD_CONFIG_FILE  = CONF_DIRECTORY + "shard-conf.json"
)

var EtcdEndpoints = []string{"http://localhost:2379"}

const (
	SERVER_PORT = "4321"
	// Define the interval for the health check
	HEALTH_CHECK_INTERVAL = 30 * time.Second
	CLEANUP_DURATION      = time.Duration(1 * time.Minute)
	DEFAULT_CTX_TIMEOUT   = 5 * time.Second
	HEALTH_CHECK_PORT     = 4322
	TRANSACTIONAL         = 0
	ACTIVE                = 1
	SNAPSHOT_DIRECTORY    = "/home/sahilsarwar/projects/kv/snapshot/"
	CONF_DIRECTORY        = "/home/sahilsarwar/projects/kv/conf/"
	PUB_SUB               = 2
	SUBSCRIBE             = "SUBSCRIBE"
	PUBLISH               = "PUBLISH"
	GET                   = "GET"
	SET                   = "SET"
	DEL                   = "DELETE"
	SET_TTL               = "SET-TTL"
	EXISTS                = "EXISTS"
	EXPIRE                = "EXPIRE"
	REPLICATE             = "REPLICATE"
	SNAPSHOT              = "SNAPSHOT"
	BEGIN                 = "BEGIN"
	COMMIT                = "COMMIT"
	ROLLBACK              = "ROLLBACK"
	SHUTDOWN              = "SHUTDOWN"
	MAKE_MASTER           = "MAKE_MASTER"
	MAKE_SLAVE            = "MAKE_SLAVE"
	CONFIG                = "CONFIG"
	PING                  = "PING"
	COMM                  = "COMM"
	KV_ETCD_ENDPOINT      = "http://localhost:2379"
	KV_ETCD_KEY           = "/kv/"
)
