package cluster

import (
	"context"
	pool "github.com/jolestar/go-commons-pool/v2"
	"go-redis/config"
	database2 "go-redis/database"
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/consistenthash"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"strings"
)

type ClusterDatabase struct {
	self string

	nodes          []string
	peerPicker     *consistenthash.NodeMap
	peerConnection map[string]*pool.ObjectPool
	db             database.Database
}

func MakeClusterDatabase() *ClusterDatabase {
	cluster := &ClusterDatabase{
		self:           config.Properties.Self,
		db:             database2.NewStandaloneDatabase(),
		peerPicker:     consistenthash.NewNodeMap(nil),
		peerConnection: make(map[string]*pool.ObjectPool),
	}
	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	for _, peer := range config.Properties.Peers {
		nodes = append(nodes, peer)
	}
	nodes = append(nodes, config.Properties.Self)
	cluster.nodes = nodes
	cluster.peerPicker.AddNode(nodes...)
	ctx := context.Background()
	for _, peer := range config.Properties.Peers {
		pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{
			Peer: peer,
		})
	}
	return cluster
}

type CmdFunc func(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply

var router = makeRouter()

func (c *ClusterDatabase) Exec(conn resp.Connection, args [][]byte) (result resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			result = &reply.UnknownErrReply{}
		}
	}()
	cmdName := strings.ToLower(string(args[0]))
	cmdFunc, ok := router[cmdName]
	if !ok {
		return reply.MakeStandardErrReply("not supported cmd")
	}
	cmdFunc(c, conn, args)
	return
}

func (c *ClusterDatabase) Close() {
	c.db.Close()
}

func (c *ClusterDatabase) AfterClientClose(conn resp.Connection) {
	c.db.AfterClientClose(conn)
}
