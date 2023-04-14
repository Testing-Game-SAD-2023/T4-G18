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


