package main

import (
	"time"
)

type ApiError struct {
	code    int
	Message string `json:"message"`
	err     error
}

func (ae ApiError) Error() string {
	return ae.Message
}

type CreateGameRequest struct {
	Name    string   `json:"name"`
	Players []string `json:"players"`
}

func (CreateGameRequest) Validate() error {
	return nil
}

type CreateRoundRequest struct {
	GameId      int64  `json:"gameId"`
	TestClassId string `json:"testClassId"`
	Order       int    `json:"order"`
}

func (CreateRoundRequest) Validate() error {
	return nil
}

type CreateTurnsRequest struct {
	RoundId int64    `json:"roundId"`
	Players []string `json:"players"`
}

func (CreateTurnsRequest) Validate() error {
	return nil
}

type UpdateRoundRequest struct {
	Order int `json:"order"`
}

func (UpdateRoundRequest) Validate() error {
	return nil
}

type UpdateTurnRequest struct {
	Scores   string `json:"scores"`
	IsWinner bool   `json:"isWinner"`
}

func (UpdateTurnRequest) Validate() error {
	return nil
}

type UpdateGameRequest struct {
	CurrentRound int    `json:"currentRound"`
	Name         string `json:"name"`
}

func (UpdateGameRequest) Validate() error {
	return nil
}

type GameDto struct {
	ID           int64     `json:"id"`
	CurrentRound int       `json:"currentRound"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	PlayersCount int       `json:"playersCount"`
	Name         string    `json:"name"`
}

type RoundDto struct {
	ID          int64     `json:"id"`
	Order       int       `json:"order"`
	TestClassId string    `json:"testClassId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type TurnDto struct {
	ID        int64     `json:"id"`
	IsWinner  bool      `json:"isWinner"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	PlayerID  int64     `json:"playerId"`
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

func mapToGameDTO(g *GameModel) *GameDto {
	return &GameDto{
		ID:           g.ID,
		CurrentRound: g.CurrentRound,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
		Name:         g.Name,
		PlayersCount: g.PlayersCount,
	}
}

func mapToRoundDTO(g *RoundModel) *RoundDto {
	return &RoundDto{
		ID:          g.ID,
		Order:       g.Order,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
		TestClassId: g.TestClassId,
	}
}

func mapToTurnDTO(t *TurnModel) *TurnDto {
	return &TurnDto{
		ID:        t.ID,
		IsWinner:  t.IsWinner,
		Scores:    t.Scores,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		PlayerID:  t.PlayerID,
	}
}
