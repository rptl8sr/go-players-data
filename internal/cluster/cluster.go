package cluster

import (
	"go-players-data/internal/model"
)

type cluster struct {
}

type Cluster interface {
	ByStoreNumber(players []*model.Player) map[int][]*model.Player
}

func New() Cluster {
	return &cluster{}
}

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
