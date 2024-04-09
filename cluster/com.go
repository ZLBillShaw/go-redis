package cluster

import (
	"context"
	"errors"
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/client"
	"go-redis/resp/reply"
	"strconv"
)

func (c *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	pool, ok := c.peerConnection[peer]
	if !ok {
		return nil, errors.New("connection not found")
	}
	object, err := pool.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}
	client2, ok := object.(*client.Client)
	if !ok {
		return nil, errors.New("connection pool wrong type")
	}
	return client2, err
}

func (c *ClusterDatabase) returnPeerClient(peer string, peerClient *client.Client) error {
	pool, ok := c.peerConnection[peer]
	if !ok {
		return errors.New("connection not found")
	}
	return pool.ReturnObject(context.Background(), peerClient)

}

func (c *ClusterDatabase) relay(peer string, conn resp.Connection, args [][]byte) resp.Reply {
	if peer == c.self {
		return c.db.Exec(conn, args)
	}
	peerClient, err := c.getPeerClient(peer)
	if err != nil {
		return reply.MakeStandardErrReply(err.Error())
	}
	defer func() {
		_ = c.returnPeerClient(peer, peerClient)
	}()

	peerClient.Send(utils.ToCmdLine("SELECT", strconv.Itoa(conn.GetDBIndex())))
	return peerClient.Send(args)
}

func (c *ClusterDatabase) broadcast(conn resp.Connection, args [][]byte) map[string]resp.Reply {
	results := make(map[string]resp.Reply)
	for _, node := range c.nodes {
		result := c.relay(node, conn, args)
		results[node] = result
	}

	return results
}
