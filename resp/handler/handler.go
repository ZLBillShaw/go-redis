package handler

import (
	"go-redis/cluster"
	"go-redis/config"
	"go-redis/database"
	"go-redis/lib/sync/atomic"
	"sync"
)

var unknownErrReplyBytes = []byte("ERR unknown\r\n")

type RespHandler struct {
	activeConn sync.Map
	db         *database.StandaloneDatabase
	dbcluster  *cluster.ClusterDatabase
	closing    atomic.Boolean
}

func MakeHandler() *RespHandler {
	if config.Properties.Self != "" && len(config.Properties.Peers) > 0 {
		clusterDB := cluster.MakeClusterDatabase()
		return &RespHandler{dbcluster: clusterDB}
	} else {
		db := database.NewStandaloneDatabase()
		return &RespHandler{db: db}
	}

}

// TODO
func (h *RespHandler) closeClient() {
}
