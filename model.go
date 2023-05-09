package main

import "time"

type GameModel struct {
	CurrentRound int   `gorm:"default:1"`
	ID           int64 `gorm:"primaryKey;autoIncrement"`
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
	ID          int64       `gorm:"primaryKey;autoIncrement"`
	Order       int         `gorm:"not null;default:1"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime"`
	CreatedAt   time.Time   `gorm:"autoCreateTime"`
	Turns       []TurnModel `gorm:"foreignKey:RoundID"`
	IdTestClass string      `gorm:"not null"`
	GameID      int64       `gorm:"not null"`
}

func (g RoundModel) TableName() string {
	return "rounds"
}

type TurnModel struct {
	ID        int64         `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time     `gorm:"autoCreateTime"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime"`
	Metadata  MetadataModel `gorm:"foreignKey:TurnID"`
	Scores    string        `gorm:"default:null"`
	IsWinner  bool          `gorm:"default:false"`
	PlayerID  int64         `gorm:"index:idx_playerturn;unique;not null"`
	RoundID   int64         `gorm:"index:idx_playerturn;unique;not null"`
}

func (t TurnModel) TableName() string {
	return "turns"
}

type PlayerModel struct {
	ID          int64             `gorm:"primaryKey;autoIncrement"`
	AccountID   string            `gorm:"unique"`
	CreatedAt   time.Time         `gorm:"autoCreateTime"`
	UpdatedAt   time.Time         `gorm:"autoUpdateTime"`
	Turns       []TurnModel       `gorm:"foreignKey:PlayerID"`
	PlayerGames []PlayerGameModel `gorm:"foreignKey:PlayerID"`
}

func (p PlayerModel) TableName() string {
	return "players"
}

type MetadataModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	TurnID    int64     `gorm:"unique;not null"`
	Path      string    `gorm:"unique;not null"`
}

func (t MetadataModel) TableName() string {
	return "metadata"
}

type PlayerGameModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	IsWinner  bool      `gorm:"default:false"`
	PlayerID  int64     `gorm:"index:idx_playergame;unique;not null;"`
	GameID    int64     `gorm:"index:idx_playergame;unique;not null"`
}

func (t PlayerGameModel) TableName() string {
	return "player_game"
}
