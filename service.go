package main

import "errors"

var (
	ErrNotFound   = errors.New("not found")
	ErrBadRequest = errors.New("bad request")
)

type GameRepository interface {
	Create(request *CreateGameRequest) (*GameModel, error)
	FindById(id uint64) (*GameModel, error)
	Delete(id uint64) error
}

type GameService struct {
	storage GameRepository
}

func NewGameController(storage GameRepository) *GameService {
	return &GameService{
		storage: storage,
	}
}

func (gc *GameService) Create(request *CreateGameRequest) (*GameModel, error) {
	return gc.storage.Create(request)
}

func (gc *GameService) FindByID(id uint64) (*GameModel, error) {
	return gc.storage.FindById(id)
}

func (gc *GameService) Delete(id uint64) error {
	return gc.storage.Delete(id)
}

type RoundRepository interface {
	Create(request *CreateRoundRequest) (*RoundModel, error)
	FindById(id uint64) (*RoundModel, error)
	Delete(id uint64) error
}

type RoundService struct {
	storage RoundRepository
}

func NewRoundController(storage RoundRepository) *RoundService {
	return &RoundService{
		storage: storage,
	}
}

func (rc *RoundService) Create(request *CreateRoundRequest) (*RoundModel, error) {
	return rc.storage.Create(request)
}

func (rc *RoundService) FindByID(id uint64) (*RoundModel, error) {
	return rc.storage.FindById(id)
}

func (rc *RoundService) Delete(id uint64) error {
	return rc.storage.Delete(id)
}
