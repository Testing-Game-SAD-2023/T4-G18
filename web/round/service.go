package round

import (
	"errors"

	"github.com/alarmfox/game-repository/model"
	"github.com/alarmfox/game-repository/web"
	"github.com/alarmfox/game-repository/web/game"
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

func (rs *Repository) Create(r *CreateRequest) (Round, error) {
	var round model.Round

	err := rs.db.Transaction(func(tx *gorm.DB) error {

		err := tx.
			Select("id").
			First(&game.Game{}, r.GameId).
			Error

		if err != nil {
			return err
		}
		var lastRound model.Round
		err = tx.Where(&model.Round{GameID: r.GameId}).
			Order(clause.OrderBy{
				Columns: []clause.OrderByColumn{
					{
						Column: clause.Column{
							Name: "order",
						},
						Desc: true,
					},
				},
			}).
			Last(&lastRound).
			Error

		round = model.Round{
			GameID:      r.GameId,
			TestClassId: r.TestClassId,
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			round.Order = lastRound.Order + 1
		} else if err != nil {
			return err
		} else {
			round.Order = 1
		}

		return tx.
			Create(&round).
			Error

	})

	return fromModel(&round), web.MakeServiceError(err)
}

func (rs *Repository) Update(id int64, r *UpdateRequest) (Round, error) {

	var (
		round model.Round
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

	return fromModel(&round), web.MakeServiceError(err)
}

func (rs *Repository) FindById(id int64) (Round, error) {
	var round model.Round

	err := rs.db.
		First(&round, id).
		Error

	return fromModel(&round), web.MakeServiceError(err)
}

func (rs *Repository) FindByGame(id int64) ([]Round, error) {
	var rounds []model.Round

	err := rs.db.
		Scopes(web.WithOrder("order")).
		Where(&model.Round{GameID: id}).
		Find(&rounds).
		Error

	resp := make([]Round, len(rounds))
	for i, round := range rounds {
		resp[i] = fromModel(&round)
	}

	return resp, web.MakeServiceError(err)
}

func (rs *Repository) Delete(id int64) error {
	err := rs.db.Transaction(func(tx *gorm.DB) error {
		var round model.Round
		db := rs.db.
			Where(&model.Round{ID: id}).
			Clauses(clause.Returning{}).
			Delete(&round)

		if db.Error != nil {
			return db.Error
		} else if db.RowsAffected < 1 {
			return web.ErrNotFound
		}

		err := rs.db.
			Model(&model.Round{}).
			Where(&model.Round{GameID: round.GameID}).
			Where("\"order\" > ?", round.Order).
			UpdateColumn("order", gorm.Expr("\"order\" - ?", 1)).
			Error

		return err
	})

	return web.MakeServiceError(err)
}
