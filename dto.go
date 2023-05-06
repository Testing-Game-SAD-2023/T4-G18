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
	PlayersCount int    `json:"playersCount"`
	Name         string `json:"name"`
}

type CreateRoundRequest struct {
	IdGame      uint64 `json:"idGame"`
	IdTestClass string `json:"idTestClass"`
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
	IdTestClass string    `json:"idTestClass"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
