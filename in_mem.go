package main

import "sync"

type GameInMem struct {
	mx     sync.Mutex
	data   map[uint64]*GameModel
	nextId uint64
}

func (gin *GameInMem) Create(r *CreateGameRequest) (*GameModel, error) {
	gin.mx.Lock()
	defer gin.mx.Unlock()

	g := &GameModel{
		ID:           gin.nextId,
		PlayersCount: r.PlayersCount,
	}

	gin.data[gin.nextId] = g
	gin.nextId += 1
	return g, nil
}
