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

func (gs *GameStorage) Create(r *CreateGameRequest) (*GameModel, error) {
	g := GameModel{
		PlayersCount: r.PlayersCount,
		Name:         r.Name,
	}
	err := gs.db.
		Create(&g).
		Error

	return &g, handleDbError(err)
}

func (gs *GameStorage) FindById(id int64) (*GameModel, error) {
	var game GameModel
	err := gs.db.
		First(&game, id).
		Error

	return &game, handleDbError(err)
}

func (gs *GameStorage) FindByInterval(i *IntervalParams, p *PaginationParams) ([]GameModel, int64, error) {
	var games []GameModel
	var n int64

	err := gs.db.
		Scopes(Intervaled(i), Paginated(p)).
		Find(&games).
		Count(&n).
		Error

	return games, n, handleDbError(err)
}

func (gs *GameStorage) FindByRound(id int64) (*GameModel, error) {

	var game GameModel

	err := gs.db.
		Preload("Rounds", &RoundModel{ID: id}).
		First(&game).
		Error

	return &game, handleDbError(err)
}

func (gs *GameStorage) Delete(id int64) error {
	db := gs.db.
		Where(&GameModel{ID: id}).
		Delete(&GameModel{})

	if db.Error != nil {
		return handleDbError(db.Error)
	} else if db.RowsAffected < 1 {
		return ErrNotFound
	}
	return nil
}

func (gs *GameStorage) Update(id int64, r *UpdateGameRequest) (*GameModel, error) {

	var game GameModel

	err := gs.db.
		Model(&game).
		Clauses(clause.Returning{}).
		Where(&GameModel{ID: id}).
		Updates(r).
		Error

	return &game, handleDbError(err)
}

type RoundStorage struct {
	db *gorm.DB
}

func NewRoundStorage(db *gorm.DB) *RoundStorage {
	return &RoundStorage{
		db: db,
	}
}

func (rs *RoundStorage) Create(r *CreateRoundRequest) (*RoundModel, error) {

	round := RoundModel{
		Order:       r.Order,
		GameID:      r.GameId,
		TestClassId: r.TestClassId,
	}
	err := rs.db.
		Create(&round).
		Error

	return &round, handleDbError(err)
}

func (rs *RoundStorage) Update(id int64, r *UpdateRoundRequest) (*RoundModel, error) {

	var round RoundModel

	err := rs.db.
		Model(&round).
		Clauses(clause.Returning{}).
		Where(&RoundModel{ID: id}).
		Updates(r).
		Error

	return &round, handleDbError(err)
}

func (rs *RoundStorage) FindById(id int64) (*RoundModel, error) {
	var round RoundModel

	err := rs.db.
		First(&round, id).
		Error

	return &round, handleDbError(err)
}

func (rs *RoundStorage) FindByGame(id int64) ([]RoundModel, error) {
	var rounds []RoundModel

	err := rs.db.
		Scopes(OrderBy("order")).
		Find(&rounds).
		Error

	return rounds, handleDbError(err)
}

func (rs *RoundStorage) Delete(id int64) error {
	return rs.db.Transaction(func(tx *gorm.DB) error {
		var round RoundModel
		db := rs.db.
			Where(&RoundModel{ID: id}).
			Clauses(clause.Returning{}).
			Delete(&round)

		if db.Error != nil {
			return handleDbError(db.Error)
		} else if db.RowsAffected < 1 {
			return ErrNotFound
		}

		err := rs.db.
			Model(&RoundModel{}).
			Where(&RoundModel{GameID: round.GameID}).
			Where("\"order\" > ?", round.Order).
			UpdateColumn("order", gorm.Expr("\"order\" - ?", 1)).
			Error

		return handleDbError(err)
	})
}

type TurnStorage struct {
	db *gorm.DB
}

func NewTurnStorage(db *gorm.DB) *TurnStorage {
	return &TurnStorage{
		db: db,
	}
}

func (ts *TurnStorage) Create(r *CreateTurnRequest) (*TurnModel, error) {

	turn := TurnModel{
		PlayerID: r.PlayerId,
		Scores:   r.Scores,
		RoundID:  r.RoundId,
	}
	err := ts.db.
		Create(&turn).
		Error

	return &turn, handleDbError(err)
}

func (ts *TurnStorage) Update(id int64, r *UpdateTurnRequest) (*TurnModel, error) {

	var turn TurnModel

	err := ts.db.
		Model(&turn).
		Clauses(clause.Returning{}).
		Where(&TurnModel{ID: id}).
		Updates(r).
		Error

	return &turn, handleDbError(err)
}

func (ts *TurnStorage) FindById(id int64) (*TurnModel, error) {
	var turn TurnModel

	err := ts.db.
		First(&turn, id).
		Error

	return &turn, handleDbError(err)
}

func (ts *TurnStorage) FindByRound(id int64) ([]TurnModel, error) {
	var turns []TurnModel

	err := ts.db.
		Where(&TurnModel{RoundID: id}).
		Find(&turns).
		Error

	return turns, handleDbError(err)
}

func (ts *TurnStorage) Delete(id int64) error {

	db := ts.db.
		Where(&TurnModel{ID: id}).
		Delete(&TurnModel{})

	if db.Error != nil {
		return db.Error
	} else if db.RowsAffected < 1 {
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
	var meta MetadataModel

	err := ms.db.Model(&meta).
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "turn_id"}},
				DoUpdates: clause.Assignments(map[string]interface{}{"path": path}),
			},
			clause.Returning{},
		).
		Create(&MetadataModel{TurnID: id, Path: path}).
		Error

	return handleDbError(err)
}

func (ms *MetadataStorage) FindByTurn(id int64) (*MetadataModel, error) {
	var meta MetadataModel

	err := ms.db.
		First(&meta, &MetadataModel{TurnID: id}).
		Error

	return &meta, handleDbError(err)
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

	err := ps.db.
		First(&player, id).
		Error

	return &player, handleDbError(err)
}
