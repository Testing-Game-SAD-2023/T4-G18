package main

import "errors"

var (
	ErrNotFound = errors.New("not found")
)

type GameRepository interface {
	Create(request *CreateGameRequest) (*GameModel, error)
	FindById(id uint64) (*GameModel, error)
	Delete(id uint64) error
}

type GameController struct {
	storage GameRepository
}

func NewGameController(storage GameRepository) *GameController {
	return &GameController{
		storage: storage,
	}
}

func (gc *GameController) Create(request *CreateGameRequest) (*GameModel, error) {
	return gc.storage.Create(request)
}

func (gc *GameController) FindByID(id uint64) (*GameModel, error) {
	return gc.storage.FindById(id)
}

func (gc *GameController) Delete(id uint64) error {
	return gc.storage.Delete(id)
}

type RoundRepository interface {
	Create(request *CreateRoundRequest) (*RoundModel, error)
	FindById(id uint64) (*RoundModel, error)
	Delete(id uint64) error
}

type RoundController struct {
	storage RoundRepository
}

func NewRoundController(storage RoundRepository) *RoundController {
	return &RoundController{
		storage: storage,
	}
}

func (rc *RoundController) Create(request *CreateRoundRequest) (*RoundModel, error) {
	return rc.storage.Create(request)
}

func (rc *RoundController) FindByID(id uint64) (*RoundModel, error) {
	return rc.storage.FindById(id)
}

func (rc *RoundController) Delete(id uint64) error {
	return rc.storage.Delete(id)
}
