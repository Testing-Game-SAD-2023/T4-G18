package main

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GameStorage struct {
	db *gorm.DB
}

func NewGameStorage(db *gorm.DB) *GameStorage {
	return &GameStorage{
		db: db,
	}
}

func (gs *GameStorage) Create(request *CreateGameRequest) (*GameModel, error) {
	g := GameModel{
		PlayersCount: request.PlayersCount,
		Name:         request.Name,
	}
	err := gs.db.Create(&g).Error

	return &g, err
}

func (gs *GameStorage) FindById(id uint64) (*GameModel, error) {
	var game GameModel
	err := gs.db.First(&game, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return &game, nil
}

func (gs *GameStorage) Delete(id uint64) error {
	rowsAffected := gs.db.Delete(&GameModel{}, id).RowsAffected
	if rowsAffected < 1 {
		return ErrNotFound
	}
	return nil
}

func (gs *GameStorage) Update(id uint64, ug *UpdateGameRequest) (*GameModel, error) {
	tx := gs.db.Begin()
	defer tx.Rollback()

	var game GameModel
	err := tx.First(&game, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	if err := gs.db.Model(&game).Updates(ug).Error; err != nil {
		return nil, err
	}

	tx.Commit()

	return &game, nil
}

type RoundStorage struct {
	db *gorm.DB
}

func NewRoundStorage(db *gorm.DB) *RoundStorage {
	return &RoundStorage{
		db: db,
	}
}

func (rs *RoundStorage) Create(request *CreateRoundRequest) (*RoundModel, error) {
	r := RoundModel{
		GameID:      request.IdGame,
		IdTestClass: request.IdTestClass,
	}
	if err := rs.db.Create(&r).Error; err != nil {
		return nil, err
	}

	return &r, nil
}

func (rs *RoundStorage) FindById(id uint64) (*RoundModel, error) {
	var round RoundModel
	err := rs.db.First(&round, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return &round, nil
}

func (rs *RoundStorage) Delete(id uint64) error {
	rowsAffected := rs.db.Delete(&RoundModel{}, id).RowsAffected
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

func (ts *TurnStorage) FindGameByTurn(id uint64) (*GameModel, error) {
	var game GameModel
	if err := ts.db.Preload("Rounds.Turns", "turn_id = ?", id).First(&game).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &game, nil

}

func (ts *TurnStorage) UpdateMetadata(id uint64, path string) error {

	meta := MetadataModel{
		TurnID: id,
		Path:   path,
	}
	return ts.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "turn_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"path": path}),
	}).Create(&meta).Error
}

func (ts *TurnStorage) FindMetadataByTurn(turnId uint64) (*MetadataModel, error) {
	var meta MetadataModel
	if err := ts.db.First(&meta, "turn_id = ?", turnId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &meta, nil
}

func (db *TurnStorage) Create(request *CreateTurnRequest) (*TurnModel, error) {
	t := TurnModel{
		PlayerID:      request.IdPlayer,
		RoundID: 	   request.IdRound,
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