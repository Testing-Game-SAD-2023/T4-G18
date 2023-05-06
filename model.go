package main

import "time"

type GameModel struct {
	CurrentRound int    `gorm:"default:1"`
	ID           uint64 `gorm:"primaryKey;autoIncrement"`
	Name         string
	CreatedAt    time.Time         `gorm:"autoCreateTime"`
	UpdatedAt    time.Time         `gorm:"autoUpdateTime"`
	Rounds       []RoundModel      `gorm:"foreignKey:GameID"`
	PlayerGame   []PlayerGameModel `gorm:"foreignKey:GameID"`
	PlayersCount int
}

func (g GameModel) TableName() string {
	return "games"
}

type RoundModel struct {
	ID          uint64      `gorm:"primaryKey;autoIncrement"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime"`
	CreatedAt   time.Time   `gorm:"autoCreateTime"`
	Turns       []TurnModel `gorm:"foreignKey:RoundID"`
	IdTestClass string
	GameID      uint64
}

func (g RoundModel) TableName() string {
	return "rounds"
}

type TurnModel struct {
	ID        uint64        `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time     `gorm:"autoCreateTime"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime"`
	Metadata  MetadataModel `gorm:"foreignKey:TurnID"`
	IsWinner  bool
	PlayerID  string
	RoundID   uint64
}

func (t TurnModel) TableName() string {
	return "turns"
}

type PlayerModel struct {
	ID          string            `gorm:"primaryKey"`
	CreatedAt   time.Time         `gorm:"autoCreateTime"`
	UpdatedAt   time.Time         `gorm:"autoUpdateTime"`
	Turns       []TurnModel       `gorm:"foreignKey:PlayerID"`
	PlayerGames []PlayerGameModel `gorm:"foreignKey:PlayerID"`
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

type PlayerGameModel struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	IsWinner  bool
	PlayerID  uint64
	GameID    uint64
}

func (t PlayerGameModel) TableName() string {
	return "player_game"
}
