package main

import "time"

type ApiError struct {
	code    int
	Message string `json:"message"`
}

func (ae ApiError) Error() string {
	return ae.Message
}

type CreateGameRequest struct {
	PlayersCount int `json:"playersCount"`
}

type CreateRoundRequest struct {
	IdGame      uint64 `json:"idGame"`
	IdTestClass string `json:"idTestClass"`
}

type CreateTurnRequest struct {
	IdPlayer     uint64 `json:"idPlayer"`
	IdRound      uint64 `json:"idRound"`
}

type UpdateGameRequest struct {
	CurrentRound int `json:"currentRound"`
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

type TurnDto struct {
	ID          uint64    `json:"id"`
	IsWinner 	bool      `json:"isWinner"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	PlayerID    uint64	  `json:"idPlayer"`
}

