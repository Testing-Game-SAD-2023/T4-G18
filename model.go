package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GameModel struct {
	CurrentRound int   `gorm:"default:1"`
	ID           int64 `gorm:"primaryKey;autoIncrement"`
	Name         string
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	PlayersCount int
	Rounds       []RoundModel      `gorm:"foreignKey:GameID;constraint:OnDelete:CASCADE;"`
	PlayerGame   []PlayerGameModel `gorm:"foreignKey:GameID;constraint:OnDelete:CASCADE;"`
}

func (GameModel) TableName() string {
	return "games"
}

type RoundModel struct {
	ID          int64       `gorm:"primaryKey;autoIncrement"`
	Order       int         `gorm:"not null;default:1"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime"`
	CreatedAt   time.Time   `gorm:"autoCreateTime"`
	Turns       []TurnModel `gorm:"foreignKey:RoundID;constraint:OnDelete:CASCADE;"`
	TestClassId string      `gorm:"not null"`
	GameID      int64       `gorm:"not null"`
}

func (RoundModel) TableName() string {
	return "rounds"
}

func (rm *RoundModel) BeforeUpdate(tx *gorm.DB) error {
	var round RoundModel

	err := tx.Where(&RoundModel{GameID: rm.GameID}).
		Clauses(clause.OrderBy{
			Columns: []clause.OrderByColumn{
				{
					Column: clause.Column{
						Table: round.TableName(),
						Name:  "order",
					},
					Desc: true,
				},
			},
		}).
		Last(&round).
		Error

	if err != nil {
		return err
	}

	r := tx.Statement.Dest.(*UpdateRoundRequest)
	if (r.Order - round.Order) != 1 {
		return fmt.Errorf("%w: last round has order %d; expected %d",
			ErrInvalidRoundOrder,
			round.Order,
			round.Order+1,
		)
	}

	return nil
}

func (rm *RoundModel) BeforeCreate(tx *gorm.DB) error {
	var round RoundModel

	err := tx.Where(&RoundModel{GameID: rm.GameID}).
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
		Last(&round).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if rm.Order == 1 {
			return nil
		}
		return fmt.Errorf("%w: first round must have order 1", ErrInvalidRoundOrder)
	} else if err != nil {
		return err
	}

	if (rm.Order - round.Order) != 1 {
		return fmt.Errorf("%w: last round has order %d; expected %d",
			ErrInvalidRoundOrder,
			round.Order,
			round.Order+1,
		)
	}

	return nil
}

type TurnModel struct {
	ID        int64         `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time     `gorm:"autoCreateTime"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime"`
	Metadata  MetadataModel `gorm:"foreignKey:TurnID;constraint:OnDelete:SET NULL;"`
	Scores    string        `gorm:"default:null"`
	IsWinner  bool          `gorm:"default:false"`
	PlayerID  int64         `gorm:"index:idx_playerturn,unique;not null"`
	RoundID   int64         `gorm:"index:idx_playerturn,unique;not null"`
}

func (TurnModel) TableName() string {
	return "turns"
}

type PlayerModel struct {
	ID          int64             `gorm:"primaryKey;autoIncrement"`
	AccountID   string            `gorm:"unique"`
	CreatedAt   time.Time         `gorm:"autoCreateTime"`
	UpdatedAt   time.Time         `gorm:"autoUpdateTime"`
	Turns       []TurnModel       `gorm:"foreignKey:PlayerID;constraint:OnDelete:SET NULL;"`
	PlayerGames []PlayerGameModel `gorm:"foreignKey:PlayerID;constraint:OnDelete:SET NULL;"`
}

func (PlayerModel) TableName() string {
	return "players"
}

type MetadataModel struct {
	ID        int64         `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time     `gorm:"autoCreateTime"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime"`
	TurnID    sql.NullInt64 `gorm:"unique"`
	Path      string        `gorm:"unique;not null"`
}

func (MetadataModel) TableName() string {
	return "metadata"
}

type PlayerGameModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	IsWinner  bool      `gorm:"default:false"`
	PlayerID  int64     `gorm:"index:idx_playergame,unique;not null;"`
	GameID    int64     `gorm:"index:idx_playergame,unique;not null"`
}

func (PlayerGameModel) TableName() string {
	return "player_game"
}


type RobotModel struct {
	ID        	int64       `gorm:"primaryKey;autoIncrement"`
	CreatedAt 	time.Time   `gorm:"autoCreateTime"`
	UpdatedAt 	time.Time   `gorm:"autoUpdateTime"`
	TestClassId string      `gorm:"not null"`
	Scores      string      `gorm:"default:null"`
	Difficulty  string      `gorm:"not null"`
	Type        int8      	`gorm:"not null"`
}

func (RobotModel) TableName() string {
	return "robots"
}
