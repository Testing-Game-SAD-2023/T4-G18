package main

import (
	"time"
)

type ApiError struct {
	code    int
	Message string `json:"message"`
}

func (ae ApiError) Error() string {
	return ae.Message
}

type CreateGameRequest struct {
	PlayersCount int    `json:"playersCount"`
	Name         string `json:"name"`
}

type CreateRoundRequest struct {
	IdGame      uint64 `json:"idGame"`
	IdTestClass string `json:"idTestClass"`
	Order       int    `json:"order"`
}

type CreateTurnRequest struct {
	IdPlayer uint64 `json:"idPlayer"`
	IdRound  uint64 `json:"idRound"`
	Scores   string `json:"scores"`
}

type UpdateRoundRequest struct {
	//TO-DO
}

type UpdateTurnRequest struct {
	Scores   string `json:"scores"`
	IsWinner bool   `json:"isWinner"`
}

type UpdateGameRequest struct {
	CurrentRound int    `json:"currentRound"`
	Name         string `json:"name"`
}

type GameDto struct {
	ID           uint64    `json:"id"`
	CurrentRound int       `json:"currentRound"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	PlayersCount int       `json:"playersCount"`
	Name         string    `json:"name"`
}

type RoundDto struct {
	ID          uint64    `json:"id"`
	Order       int       `json:"order"`
	IdTestClass string    `json:"idTestClass"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type TurnDto struct {
	ID        uint64    `json:"id"`
	IsWinner  bool      `json:"isWinner"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	PlayerID  uint64    `json:"idPlayer"`
	Scores    string    `json:"scores"`
}

type PaginatedResponse struct {
	Data     any                `json:"data"`
	Metadata PaginationMetadata `json:"metadata"`
}

type PaginationMetadata struct {
	HasNext  bool  `json:"hasNext"`
	Count    int64 `json:"count"`
	Page     int64 `json:"page"`
	PageSize int64 `json:"pageSize"`
}
