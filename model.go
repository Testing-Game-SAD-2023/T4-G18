package main

import "time"

type GameModel struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement"`
	CurrentRound int
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	PlayersCount int
	Rounds       []RoundModel `gorm:"foreignKey:GameID"`
}

func (g GameModel) TableName() string {
	return "games"
}

type RoundModel struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement"`
	IdTestClass string
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	GameID      uint64
}

func (g RoundModel) TableName() string {
	return "rounds"
}

type CreateGameRequest struct {
	PlayersCount int `json:"playersCount"`
}

type CreateRoundRequest struct {
	IdGame      uint64 `json:"idGame"`
	IdTestClass string `json:"idTestClass"`
}

type GameDto struct {
	ID           uint64    `json:"id"`
	CurrentRound int       `json:"currentRound"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	PlayersCount int       `json:"playersCount"`
}

type RoundDto struct {
	ID          uint64    `json:"id"`
	IdTestClass string    `json:"idTestClass"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
