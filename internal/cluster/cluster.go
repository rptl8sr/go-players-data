package cluster

import (
	"go-players-data/internal/model"
)

// cluster is an unexported type implementing the Cluster interface for grouping and managing players by store numbers.
type cluster struct {
}

// Cluster defines an interface for grouping players by their store number.
type Cluster interface {
	ByStoreNumber(players []*model.Player) map[int][]*model.Player
}

// New creates a new Cluster instance.
func New() Cluster {
	return &cluster{}
}

// ByStoreNumber groups players by their store number.
// Returns a map where the key is the store number, and the value is a slice of players.
func (c *cluster) ByStoreNumber(players []*model.Player) map[int][]*model.Player {
	byStoreNumber := make(map[int][]*model.Player)

	for _, p := range players {
		if _, ok := byStoreNumber[p.StoreNumber]; !ok {
			byStoreNumber[p.StoreNumber] = []*model.Player{}
		}

		byStoreNumber[p.StoreNumber] = append(byStoreNumber[p.StoreNumber], p)
	}

	return byStoreNumber
}
