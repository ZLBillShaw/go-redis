package database

import (
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/lib/wildcard"
	"go-redis/resp/reply"
)

// DEL
func execDel(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}
	deleted := db.Removes(keys...)
	if deleted > 0 {
		db.addAof(utils.ToCmdLine2("del", args...))
	}
	return reply.MakeIntReply(int64(deleted))
}

// EXISTS
func execExists(db *DB, args [][]byte) resp.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		_, exists := db.GetEntity(key)
		if exists {
			result++
		}
	}
	return reply.MakeIntReply(result)
}

// FLUSHDB
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()
	db.addAof(utils.ToCmdLine2("flushdb", args...))
	return reply.MakeOkReply()
}

// KEYS
func execKeys(db *DB, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0]))
	result := make([][]byte, 0)
	db.data.ForEach(func(key string, val interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}

// TYPE
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeStatusReply("none")
	}
	switch entity.Data.(type) {
	case []byte:
		return reply.MakeStatusReply("string")
	}
	// TODO：other types to be implemented
	return &reply.UnknownErrReply{}
}

// RENAME
func execRename(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])
	entity, exists := db.GetEntity(src)
	if !exists {
		return reply.MakeStandardErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	db.addAof(utils.ToCmdLine2("rename", args...))
	return reply.MakeOkReply()
}

// RENAMENX
func execRenamenx(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])

	_, ok := db.GetEntity(dest)
	if ok {
		return reply.MakeIntReply(0)
	}

	entity, exists := db.GetEntity(src)
	if !exists {
		return reply.MakeStandardErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	db.addAof(utils.ToCmdLine2("renamenx", args...))
	return reply.MakeIntReply(1)
}

func init() {
	RegisterCommand("del", execDel, -2)
	RegisterCommand("exists", execExists, -2)
	RegisterCommand("flushdb", execFlushDB, -1)
	RegisterCommand("type", execType, 2) //TYPE Key0
	RegisterCommand("rename", execRename, 3)
	RegisterCommand("renamenx", execRenamenx, 3)
	RegisterCommand("keys", execKeys, 2)
}
