package main

import (
	"errors"

	"gorm.io/gorm"
)

type GameStorage struct {
	db *gorm.DB
}

func NewGameStorage(db *gorm.DB) *GameStorage {
	return &GameStorage{
		db: db,
	}
}

func (db *GameStorage) Create(request *CreateGameRequest) (*GameModel, error) {
	g := GameModel{
		PlayersCount: request.PlayersCount,
	}
	err := db.db.Create(&g).Error

	return &g, err
}

func (db *GameStorage) FindById(id uint64) (*GameModel, error) {
	var game GameModel
	err := db.db.First(&game, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return &game, nil
}

func (db *GameStorage) Delete(id uint64) error {
	rowsAffected := db.db.Delete(&GameModel{}, id).RowsAffected
	if rowsAffected < 1 {
		return ErrNotFound
	}
	return nil
}

type RoundStorage struct {
	db *gorm.DB
}

func NewRoundStorage(db *gorm.DB) *RoundStorage {
	return &RoundStorage{
		db: db,
	}
}

func (db *RoundStorage) Create(request *CreateRoundRequest) (*RoundModel, error) {
	r := RoundModel{
		GameID:      request.IdGame,
		IdTestClass: request.IdTestClass,
	}
	err := db.db.Create(&r).Error

	return &r, err
}

func (db *RoundStorage) FindById(id uint64) (*RoundModel, error) {
	var round RoundModel
	err := db.db.First(&round, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return &round, nil
}

func (db *RoundStorage) Delete(id uint64) error {
	rowsAffected := db.db.Delete(&RoundModel{}, id).RowsAffected
	if rowsAffected < 1 {
		return ErrNotFound
	}
	return nil
}


type TurnStorage struct {
	db *gorm.DB
}

func NewTurnStorage(db *gorm.DB) *TurnStorage {
	return &TurnStorage{
		db: db,
	}
}

func (db *TurnStorage) Create(request *CreateTurnRequest) (*TurnModel, error) {
	t := TurnModel{
		PlayerID:      request.IdPlayer,
	}
	err := db.db.Create(&t).Error

	return &t, err
}

func (db *TurnStorage) FindById(id uint64) (*TurnModel, error) {
	var turn TurnModel
	err := db.db.First(&turn, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return &turn, nil
}

func (db *TurnStorage) Delete(id uint64) error {
	rowsAffected := db.db.Delete(&TurnModel{}, id).RowsAffected
	if rowsAffected < 1 {
		return ErrNotFound
	}
	return nil
}