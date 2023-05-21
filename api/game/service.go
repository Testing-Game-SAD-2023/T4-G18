package game

import (
	"github.com/alarmfox/game-repository/api"
	"github.com/alarmfox/game-repository/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
			Name: r.Name,
		}
	)
	// detect duplication in player
	if api.Duplicated(r.Players) {
		return Game{}, api.ErrInvalidParam
	}

	err := gs.db.Transaction(func(tx *gorm.DB) error {
		var (
			err         error
			players     []model.Player
			playerGames []model.PlayerGame = make([]model.PlayerGame, len(r.Players))
		)

		err = tx.
			Create(&game).
			Error

		if err != nil {
			return err
		}

		toCreate := make([]model.Player, len(r.Players))
		for i, account := range r.Players {
			toCreate[i] = model.Player{
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
			playerGames[i] = model.PlayerGame{
				GameID:   game.ID,
				PlayerID: player.ID,
			}
		}

		// create player instance in game
		return tx.Create(playerGames).Error
	})

	return fromModel(&game), api.MakeServiceError(err)
}

func (gs *Repository) FindById(id int64) (Game, error) {
	var game model.Game
	err := gs.db.
		First(&game, id).
		Error

	return fromModel(&game), api.MakeServiceError(err)
}

func (gs *Repository) FindByInterval(i *api.IntervalParams, p *api.PaginationParams) ([]Game, int64, error) {
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
