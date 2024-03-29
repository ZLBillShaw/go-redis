package aof

import (
	"go-redis/config"
	"go-redis/interface/database"
	"go-redis/lib/logger"
	"go-redis/lib/utils"
	"go-redis/resp/connection"
	"go-redis/resp/parser"
	"go-redis/resp/reply"
	"io"
	"os"
	"strconv"
)

type CmdLine = [][]byte

const aofBufferSize = 1 << 16

type payLoad struct {
	cmdLine CmdLine
	dbIndex int
}

type AofHandler struct {
	database    database.Database
	aofChan     chan *payLoad
	aofFile     *os.File
	aofFileName string
	currentDB   int
}

func NewAofHandler(database database.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFileName = config.Properties.AppendFilename
	handler.database = database
	handler.LoadAof()
	aofFile, err := os.OpenFile(handler.aofFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile
	handler.aofChan = make(chan *payLoad, aofBufferSize)
	go func() {
		handler.handleAof()
	}()
	return handler, nil
}

func (handler *AofHandler) AddAof(dbIndex int, cmd CmdLine) {
	if config.Properties.AppendOnly && handler.aofChan != nil {
		handler.aofChan <- &payLoad{
			cmdLine: cmd,
			dbIndex: dbIndex,
		}
	}
}

// HandleAof payload(set k v) <- aofChan
func (handler *AofHandler) handleAof() {
	handler.currentDB = 0
	for p := range handler.aofChan {
		if p.dbIndex != handler.currentDB {
			data := reply.MakeMultiBulkReply(utils.ToCmdLine("select", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := handler.aofFile.Write(data)
			if err != nil {
				logger.Error(err)
				continue
			}
			handler.currentDB = p.dbIndex
		}
		data := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Error(err)
			continue
		}
	}
}

func (handler *AofHandler) LoadAof() {
	file, err := os.Open(handler.aofFileName)
	if err != nil {
		logger.Error(err)
		return
	}
	defer file.Close()
	ch := parser.ParseStream(file)
	fackConn := &connection.Connection{}
	for payload := range ch {
		if payload.Err != nil {
			if payload.Err == io.EOF {
				break
			}
			logger.Error(payload.Err)
			continue
		}
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("exec err, need multi bulk")
			continue
		}
		rep := handler.database.Exec(fackConn, r.Args)
		if reply.IsErrReply(rep) {
			logger.Error(rep)
		}
	}
}
