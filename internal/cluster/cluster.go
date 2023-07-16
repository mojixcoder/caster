package cluster

import (
	"errors"
	"hash"
	"hash/fnv"
	"sync"

	"github.com/mojixcoder/caster/internal/app"
	"github.com/mojixcoder/caster/internal/config"
	"go.uber.org/zap"
)

// Node is a member of cluster.
type Node struct {
	// address is node's address.
	address string

	// isLocal determines if this is the local node or not.
	isLocal bool
}

// Cluster does the cluster managament.
type Cluster struct {
	// nodeMap maps indexes => nodes.
	nodeMap map[int]Node

	// pool is a pool of hashes.
	pool *sync.Pool
}

// IsLocal determines if this node is the local node or not.
func (n Node) IsLocal() bool {
	return n.isLocal
}

// Address returns node's address.
// It's not used if node is a local node.
func (n Node) Address() string {
	return n.address
}

// UpdateNodeMap updates the cluster's node map.
func (c *Cluster) UpdateNodeMap(nodes []config.NodeConfig) {
	nodeMap := make(map[int]Node, len(nodes))
	for _, v := range nodes {
		nodeMap[v.Index] = Node{address: v.Address, isLocal: v.IsLocal}
	}
	c.nodeMap = nodeMap
}

// ValidateNodeMap validates node map.
func (c Cluster) ValidateNodeMap() error {
	var localCount int

	for _, v := range c.nodeMap {
		if v.IsLocal() {
			localCount++
		}
	}

	switch localCount {
	case 0:
		return errors.New("no local node specified")
	case 1:
		return nil
	default:
		return errors.New("more than a local node specified")
	}
}

// GetNodeFromKey gets the node which the operation should be performed on.
func (c Cluster) GetNodeFromKey(key string) Node {
	if len(c.nodeMap) == 1 {
		return c.nodeMap[0]
	}

	hash := c.pool.Get().(hash.Hash32)
	defer func() {
		hash.Reset()
		c.pool.Put(hash)
	}()

	hash.Write([]byte(key))

	sum := hash.Sum32()
	sum %= uint32(len(c.nodeMap))

	return c.nodeMap[int(sum)]
}

// NewCluster returns a new memberlist cluster.
func NewCluster() (Cluster, error) {
	var cluster Cluster

	app.App.Logger.Info("discovering cluster nodes")

	if len(app.App.Config.Nodes) >= 1 {
		cluster.UpdateNodeMap(app.App.Config.Nodes)
	} else {
		cluster.UpdateNodeMap([]config.NodeConfig{{Address: "local_node", IsLocal: true}})
	}

	if err := cluster.ValidateNodeMap(); err != nil {
		return Cluster{}, err
	}

	for _, v := range cluster.nodeMap {
		if v.IsLocal() {
			app.App.Logger.Info("discovered node", zap.String("node", "local node"))
		} else {
			app.App.Logger.Info("discovered node", zap.String("node", v.Address()))
		}
	}

	cluster.pool = new(sync.Pool)
	cluster.pool.New = func() any {
		return fnv.New32a()
	}

	return cluster, nil
}
