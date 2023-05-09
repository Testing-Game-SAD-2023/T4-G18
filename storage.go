package main

import (
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

func (gs *GameStorage) FindById(id int64) (*GameModel, error) {
	var game GameModel
	err := gs.db.First(&game, id).Error
	if err != nil {
		return nil, handleDbError(err)
	}
	return &game, nil
}

func (gs *GameStorage) FindByInterval(i *IntervalParams, p *PaginationParams) ([]GameModel, int64, error) {
	var games []GameModel
	var n int64
	tx := gs.db.Begin()
	defer tx.Rollback()
	if err := tx.Model(&GameModel{}).Count(&n).Error; err != nil {
		return nil, 0, handleDbError(err)
	}
	if err := tx.Scopes(PaginateScope(p), IntervalScope(i)).Find(&games).Error; err != nil {
		return nil, 0, handleDbError(err)
	}
	tx.Commit()
	return games, n, nil
}

func (gs *GameStorage) FindByRound(id int64) (*GameModel, error) {

	var game GameModel
	if err := gs.db.Preload("Rounds", "id = ?", id).First(&game).Error; err != nil {
		return nil, ErrNotFound
	}

	return &game, nil
}

func (gs *GameStorage) Delete(id int64) error {
	rowsAffected := gs.db.Delete(&GameModel{}, id).RowsAffected
	if rowsAffected < 1 {
		return ErrNotFound
	}
	return nil
}

func (gs *GameStorage) Update(id int64, ug *UpdateGameRequest) (*GameModel, error) {
	tx := gs.db.Begin()
	defer tx.Rollback()

	var game GameModel

	if err := tx.First(&game, id).Error; err != nil {
		return nil, handleDbError(err)
	}

	if err := gs.db.Model(&game).Updates(ug).Error; err != nil {
		return nil, handleDbError(err)
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
		GameID:      request.GameId,
		TestClassId: request.TestClassId,
		Order:       request.Order,
	}
	if err := rs.db.Create(&r).Error; err != nil {
		return nil, handleDbError(err)
	}

	return &r, nil
}

func (rs *RoundStorage) Update(id int64, ug *UpdateRoundRequest) (*RoundModel, error) {
	tx := rs.db.Begin()
	defer tx.Rollback()

	var round RoundModel

	if err := tx.First(&round, id).Error; err != nil {
		return nil, handleDbError(err)
	}

	if err := rs.db.Model(&round).Updates(ug).Error; err != nil {
		return nil, handleDbError(err)
	}

	tx.Commit()

	return &round, nil
}

func (rs *RoundStorage) FindById(id int64) (*RoundModel, error) {
	var round RoundModel

	if err := rs.db.First(&round, id).Error; err != nil {
		return nil, handleDbError(err)
	}
	return &round, nil
}

func (rs *RoundStorage) FindByGame(id int64) ([]RoundModel, error) {
	var rounds []RoundModel

	if err := rs.db.Find(&rounds).Error; err != nil {
		return nil, handleDbError(err)
	}
	return rounds, nil
}

func (rs *RoundStorage) Delete(id int64) error {
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

func (ts *TurnStorage) Create(request *CreateTurnRequest) (*TurnModel, error) {
	t := TurnModel{
		PlayerID: request.PlayerId,
		RoundID:  request.RoundId,
		Scores:   request.Scores,
	}
	err := ts.db.Create(&t).Error

	return &t, err
}

func (ts *TurnStorage) Update(id int64, request *UpdateTurnRequest) (*TurnModel, error) {
	tx := ts.db.Begin()
	defer tx.Rollback()

	var turn TurnModel

	if err := tx.First(&turn, id).Error; err != nil {
		return nil, handleDbError(err)
	}

	if err := ts.db.Model(&turn).Updates(request).Error; err != nil {
		return nil, err
	}

	tx.Commit()

	return &turn, nil
}

func (ts *TurnStorage) FindById(id int64) (*TurnModel, error) {
	var turn TurnModel

	if err := ts.db.First(&turn, id).Error; err != nil {
		return nil, handleDbError(err)
	}
	return &turn, nil
}

func (ts *TurnStorage) FindByRound(id int64) ([]TurnModel, error) {
	var turns []TurnModel

	if err := ts.db.Where(&TurnModel{RoundID: id}).Find(&turns).Error; err != nil {
		return nil, handleDbError(err)
	}
	return turns, nil
}

func (ts *TurnStorage) Delete(id int64) error {
	rowsAffected := ts.db.Delete(&TurnModel{}, id).RowsAffected
	if rowsAffected < 1 {
		return ErrNotFound
	}
	return nil
}

type MetadataStorage struct {
	db *gorm.DB
}

func NewMetadataStorage(db *gorm.DB) *MetadataStorage {
	return &MetadataStorage{
		db: db,
	}
}

func (ms *MetadataStorage) Upsert(id int64, path string) error {

	meta := MetadataModel{
		TurnID: id,
		Path:   path,
	}

	return handleDbError(
		ms.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "turn_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"path": path}),
		}).Create(&meta).Error,
	)
}

func (ms *MetadataStorage) FindByTurn(id int64) (*MetadataModel, error) {
	var meta MetadataModel
	if err := ms.db.First(&meta, "turn_id = ?", id).Error; err != nil {
		return nil, handleDbError(err)
	}

	return &meta, nil
}

type PlayerStorage struct {
	db *gorm.DB
}

func NewPlayerStorage(db *gorm.DB) *PlayerStorage {
	return &PlayerStorage{
		db: db,
	}
}

func (ps *PlayerStorage) FindById(id int64) (*PlayerModel, error) {
	var player PlayerModel

	if err := ps.db.First(&player, id).Error; err != nil {
		return nil, handleDbError(err)
	}
	return &player, nil
}
