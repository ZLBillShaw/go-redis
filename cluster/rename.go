package cluster

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

func Rename(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	if len(cmdArgs) != 3 {
		return reply.MakeStandardErrReply("ERR wrong number of arguments for 'rename' command")
	}
	src := string(cmdArgs[1])
	dest := string(cmdArgs[2])

	srcPeer := cluster.peerPicker.PickNode(src)
	destPeer := cluster.peerPicker.PickNode(dest)

	if srcPeer != destPeer {
		return reply.MakeStandardErrReply("ERR rename must be the same peer")
	}

	return cluster.relay(srcPeer, c, cmdArgs)
}
