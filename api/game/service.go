package game

import (
	"github.com/alarmfox/game-repository/api"
	"github.com/alarmfox/game-repository/model"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (gs *Repository) Create(r *CreateRequest) (Game, error) {
	var (
		game = model.Game{
			Name:      r.Name,
			StartedAt: r.StartedAt,
			ClosedAt:  r.ClosedAt,
			Players:   make([]model.Player, len(r.Players)),
		}
	)
	// detect duplication in player
	if api.Duplicated(r.Players) {
		return Game{}, api.ErrInvalidParam
	}

	for i, player := range r.Players {
		game.Players[i] = model.Player{
			AccountID: player,
		}
	}
	err := gs.db.Transaction(func(tx *gorm.DB) error {
		return gs.db.Create(&game).Error
	})

	if err != nil {
		return Game{}, api.MakeServiceError(err)
	}

	return fromModel(&game), nil
}

func (gs *Repository) FindById(id int64) (Game, error) {
	var game model.Game
	err := gs.db.
		Preload("Players").
		First(&game, id).
		Error

	return fromModel(&game), api.MakeServiceError(err)
}

func (gs *Repository) FindByInterval(i api.IntervalParams, p api.PaginationParams) ([]Game, int64, error) {
	var games []model.Game
	var n int64

	err := gs.db.
		Scopes(api.WithInterval(i), api.WithPagination(p)).
		Find(&games).
		Count(&n).
		Error
	res := make([]Game, len(games))
	for i, game := range games {
		res[i] = fromModel(&game)
	}
	return res, n, api.MakeServiceError(err)
}

func (gs *Repository) Delete(id int64) error {
	db := gs.db.
		Where(&Game{ID: id}).
		Delete(&Game{})

	if db.Error != nil {
		return api.MakeServiceError(db.Error)
	} else if db.RowsAffected < 1 {
		return api.ErrNotFound
	}
	return nil
}

func (gs *Repository) Update(id int64, r *UpdateRequest) (Game, error) {

	var (
		game model.Game
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

	return fromModel(&game), api.MakeServiceError(err)
}

func (gr *Repository) FindByPlayer(accountId string, pp api.PaginationParams) ([]Game, int64, error) {
	var (
		// player model.Player
		count int64
		games []model.Game
	)

	err := gr.db.Transaction(func(tx *gorm.DB) error {

		association := tx.Model(&model.Player{AccountID: accountId}).
			Scopes(api.WithPagination(pp)).
			Order("created_at desc").
			Association("Games")

		count = association.Count()
		return association.Find(&games)
	})

	resp := make([]Game, len(games))
	for i, game := range games {
		resp[i] = fromModel(&game)
	}
	return resp, count, api.MakeServiceError(err)
}
