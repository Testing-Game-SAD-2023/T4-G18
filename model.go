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
	Turns       []TurnModel `gorm:"foreignKey:RoundID"`
}

func (g RoundModel) TableName() string {
	return "rounds"
}

type TurnModel struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	IsWinner  bool
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	PlayerID  string
	Metadata  MetadataModel `gorm:"foreignKey:TurnID"`
	RoundID   uint64
}

func (t TurnModel) TableName() string {
	return "turns"
}

type PlayerModel struct {
	ID        string `gorm:"primaryKey"`
	IsWinner  bool
	CreatedAt time.Time   `gorm:"autoCreateTime"`
	UpdatedAt time.Time   `gorm:"autoUpdateTime"`
	Turns     []TurnModel `gorm:"foreignKey:PlayerID"`
}

func (p PlayerModel) TableName() string {
	return "players"
}

type MetadataModel struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	TurnID    uint64    `gorm:"unique"`
	Path      string
}

func (t MetadataModel) TableName() string {
	return "metadata"
}
