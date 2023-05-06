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

func (gs *GameStorage) FindByInterval(i *IntervalParams, p *PaginationParams) ([]GameModel, error) {
	var games []GameModel
	if err := gs.db.Scopes(PaginateScope(p), IntervalScope(i)).Find(&games).Error; err != nil {
		return nil, err
	}
	return games, nil
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
		Order:       request.Order,
	}
	if err := rs.db.Create(&r).Error; err != nil {
		return nil, err
	}

	return &r, nil
}

func (rs *RoundStorage) Update(id uint64, ug *UpdateRoundRequest) (*RoundModel, error) {
	tx := rs.db.Begin()
	defer tx.Rollback()

	var round RoundModel
	err := tx.First(&round, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	if err := rs.db.Model(&round).Updates(ug).Error; err != nil {
		return nil, err
	}

	tx.Commit()

	return &round, nil
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

func (rs *RoundStorage) FindByGame(id uint64) ([]RoundModel, error) {
	var rounds []RoundModel

	if err := rs.db.Find(&rounds).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return rounds, nil
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

func (ts *TurnStorage) Create(request *CreateTurnRequest) (*TurnModel, error) {
	t := TurnModel{
		PlayerID: request.IdPlayer,
		RoundID:  request.IdRound,
		Scores:   request.Scores,
	}
	err := ts.db.Create(&t).Error

	return &t, err
}

func (ts *TurnStorage) Update(id uint64, request *UpdateTurnRequest) (*TurnModel, error) {
	tx := ts.db.Begin()
	defer tx.Rollback()

	var turn TurnModel
	err := tx.First(&turn, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	if err := ts.db.Model(&turn).Updates(request).Error; err != nil {
		return nil, err
	}

	tx.Commit()

	return &turn, nil
}

func (ts *TurnStorage) FindById(id uint64) (*TurnModel, error) {
	var turn TurnModel
	err := ts.db.First(&turn, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return &turn, nil
}

func (ts *TurnStorage) FindByRound(id uint64) ([]TurnModel, error) {
	var turns []TurnModel

	if err := ts.db.Where(&TurnModel{RoundID: id}).Find(&turns).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return turns, nil
}

func (ts *TurnStorage) Delete(id uint64) error {
	rowsAffected := ts.db.Delete(&TurnModel{}, id).RowsAffected
	if rowsAffected < 1 {
		return ErrNotFound
	}
	return nil
}
