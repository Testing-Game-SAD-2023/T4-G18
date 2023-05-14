package main

import (
	"archive/zip"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) *GameRepository {
	return &GameRepository{
		db: db,
	}
}

func duplicated(v []string) bool {
	unique := make(map[string]struct{}, len(v))
	for _, item := range v {
		if _, seen := unique[item]; seen {
			return true
		}
		unique[item] = struct{}{}
	}
	return false
}

func (gs *GameRepository) Create(r *CreateGameRequest) (*GameModel, error) {
	var (
		game GameModel = GameModel{
			Name: r.Name,
		}
	)
	// detect duplication in player
	if duplicated(r.Players) {
		return nil, ErrInvalidPlayerList
	}

	err := gs.db.Transaction(func(tx *gorm.DB) error {
		var (
			err         error
			players     []PlayerModel
			playerGames []PlayerGameModel = make([]PlayerGameModel, len(r.Players))
		)

		err = tx.
			Create(&game).
			Error

		if err != nil {
			return err
		}

		toCreate := make([]PlayerModel, len(r.Players))
		for i, account := range r.Players {
			toCreate[i] = PlayerModel{
				AccountID: account,
			}
		}

		// account creation (if not exist)
		err = tx.
			Clauses(
				clause.OnConflict{
					DoNothing: true,
				},
			).
			Create(&toCreate).
			Error

		if err != nil {
			return err
		}

		// get all players for game
		err = tx.
			Where("account_id IN ?", r.Players).
			Find(&players).
			Error

		if err != nil {
			return err
		}

		for i, player := range players {
			playerGames[i] = PlayerGameModel{
				GameID:   game.ID,
				PlayerID: player.ID,
			}
		}

		// create player instance in game
		return tx.Create(playerGames).Error
	})

	return &game, handleError(err)
}

func (gs *GameRepository) FindById(id int64) (*GameModel, error) {
	var game GameModel
	err := gs.db.
		First(&game, id).
		Error

	return &game, handleError(err)
}

func (gs *GameRepository) FindByInterval(i *IntervalParams, p *PaginationParams) ([]GameModel, int64, error) {
	var games []GameModel
	var n int64

	err := gs.db.
		Scopes(WithInterval(i), WithPagination(p)).
		Find(&games).
		Count(&n).
		Error

	return games, n, handleError(err)
}

func (gs *GameRepository) Delete(id int64) error {
	db := gs.db.
		Where(&GameModel{ID: id}).
		Delete(&GameModel{})

	if db.Error != nil {
		return handleError(db.Error)
	} else if db.RowsAffected < 1 {
		return ErrNotFound
	}
	return nil
}

func (gs *GameRepository) Update(id int64, r *UpdateGameRequest) (*GameModel, error) {

	var (
		game GameModel
		err  error
	)

	err = gs.db.Transaction(func(tx *gorm.DB) error {

		err := tx.
			First(&game, id).
			Error

		if err != nil {
			return err
		}

		return tx.Model(&game).Updates(r).Error

	})

	return &game, handleError(err)
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
	var round RoundModel

	err := rs.db.Transaction(func(tx *gorm.DB) error {

		err := tx.
			Select("id").
			First(&GameModel{}, r.GameId).
			Error

		if err != nil {
			return err
		}

		round = RoundModel{
			Order:       r.Order,
			GameID:      r.GameId,
			TestClassId: r.TestClassId,
		}

		return tx.
			Create(&round).
			Error

	})

	return &round, handleError(err)
}

func (rs *RoundStorage) Update(id int64, r *UpdateRoundRequest) (*RoundModel, error) {

	var (
		round RoundModel
		err   error
	)
	err = rs.db.Transaction(func(tx *gorm.DB) error {

		err := tx.
			First(&round, id).
			Error

		if err != nil {
			return err
		}

		return tx.Model(&round).Updates(r).Error

	})

	return &round, handleError(err)
}

func (rs *RoundStorage) FindById(id int64) (*RoundModel, error) {
	var round RoundModel

	err := rs.db.
		First(&round, id).
		Error

	return &round, handleError(err)
}

func (rs *RoundStorage) FindByGame(id int64) ([]RoundModel, error) {
	var rounds []RoundModel

	err := rs.db.
		Scopes(WithOrder("order")).
		Find(&rounds).
		Error

	return rounds, handleError(err)
}

func (rs *RoundStorage) Delete(id int64) error {
	err := rs.db.Transaction(func(tx *gorm.DB) error {
		var round RoundModel
		db := rs.db.
			Where(&RoundModel{ID: id}).
			Clauses(clause.Returning{}).
			Delete(&round)

		if db.Error != nil {
			return db.Error
		} else if db.RowsAffected < 1 {
			return ErrNotFound
		}

		err := rs.db.
			Model(&RoundModel{}).
			Where(&RoundModel{GameID: round.GameID}).
			Where("\"order\" > ?", round.Order).
			UpdateColumn("order", gorm.Expr("\"order\" - ?", 1)).
			Error

		return err
	})

	return handleError(err)
}

type TurnStorage struct {
	db      *gorm.DB
	dataDir string
}

func NewTurnRepository(db *gorm.DB, dataDir string) *TurnStorage {
	return &TurnStorage{
		db:      db,
		dataDir: dataDir,
	}
}

func (tr *TurnStorage) CreateBulk(r *CreateTurnsRequest) ([]TurnModel, error) {
	turns := make([]TurnModel, len(r.Players))

	err := tr.db.Transaction(func(tx *gorm.DB) error {
		var (
			err error
		)

		err = tx.Where(&RoundModel{ID: r.RoundId}).
			First(&RoundModel{}).
			Error
		if err != nil {
			return err
		}

		var ids []int64
		err = tx.
			Model(&PlayerModel{}).
			Select("id").
			Where("account_id in ?", r.Players).
			Find(&ids).
			Error

		if err != nil {
			return err
		}

		if len(ids) != len(r.Players) && !duplicated(r.Players) {
			return ErrInvalidPlayerList
		}

		for i, id := range ids {
			turns[i] = TurnModel{
				PlayerID: id,
				RoundID:  r.RoundId,
			}
		}

		return tx.Create(&turns).Error
	})

	return turns, handleError(err)
}

func (tr *TurnStorage) Update(id int64, r *UpdateTurnRequest) (*TurnModel, error) {

	var (
		turn TurnModel
		err  error
	)

	err = tr.db.Transaction(func(tx *gorm.DB) error {

		err := tx.
			First(&turn, id).
			Error

		if err != nil {
			return err
		}

		return tx.Model(&turn).Updates(r).Error

	})

	return &turn, handleError(err)
}

func (tr *TurnStorage) FindById(id int64) (*TurnModel, error) {
	var turn TurnModel

	err := tr.db.
		First(&turn, id).
		Error

	return &turn, handleError(err)
}

func (tr *TurnStorage) FindByRound(id int64) ([]TurnModel, error) {
	var turns []TurnModel

	err := tr.db.
		Where(&TurnModel{RoundID: id}).
		Find(&turns).
		Error

	return turns, handleError(err)
}

func (tr *TurnStorage) Delete(id int64) error {

	db := tr.db.
		Where(&TurnModel{ID: id}).
		Delete(&TurnModel{})

	if db.Error != nil {
		return db.Error
	} else if db.RowsAffected < 1 {
		return ErrNotFound
	}

	return nil

}

func (ts *TurnStorage) SaveFile(id int64, r io.Reader) error {
	err := ts.db.Transaction(func(tx *gorm.DB) error {
		var (
			err   error
			round RoundModel
		)

		err = tx.
			Joins("join turns on turns.round_id = rounds.id where turns.id  = ?", id).
			First(&round).
			Error

		if err != nil {
			return err
		}

		dst, err := os.CreateTemp("", "")
		if err != nil {
			return err
		}
		defer os.Remove(dst.Name())
		if _, err := io.Copy(dst, r); err != nil {
			return err
		}

		if zfile, err := zip.OpenReader(dst.Name()); err != nil {
			return ErrNotAZip
		} else {
			zfile.Close()
		}

		year := time.Now().Year()

		fname := path.Join(ts.dataDir,
			strconv.FormatInt(int64(year), 10),
			strconv.FormatInt(round.GameID, 10),
			fmt.Sprintf("%d.zip", id),
		)

		dir := path.Dir(fname)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}

		if err := os.Rename(dst.Name(), fname); err != nil {
			return err
		}

		return tx.FirstOrCreate(
			&MetadataModel{
				TurnID: sql.NullInt64{Int64: id, Valid: true},
				Path:   fname,
			}).
			Error

	})

	return handleError(err)

}

func (ts *TurnStorage) GetFile(id int64) (string, *os.File, error) {
	var (
		metadata MetadataModel
		err      error
	)

	err = ts.db.
		Where(&MetadataModel{TurnID: sql.NullInt64{Int64: id, Valid: true}}).
		First(&metadata).
		Error

	if err != nil {
		return "", nil, handleError(err)
	}

	f, err := os.Open(metadata.Path)

	if err != nil {
		return "", nil, err
	}

	return filepath.Base(metadata.Path), f, nil
}
