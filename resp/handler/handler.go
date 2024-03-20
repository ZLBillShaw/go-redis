package handler

import (
	"go-redis/database"
	"go-redis/lib/sync/atomic"
	"sync"
)

var unknownErrReplyBytes = []byte("ERR unknown\r\n")

type RespHandler struct {
	activeConn sync.Map
	db         *database.Database
	closing    atomic.Boolean
}

func MakeHandler() *RespHandler {
	db := database.NewDatabase()
	return &RespHandler{db: db}
}

// TODO
func (h *RespHandler) closeClient() {
}
