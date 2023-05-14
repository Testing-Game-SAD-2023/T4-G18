package main

import (
	"strconv"
	"time"
)

// ApiError represents the http error returned by the REST service.
// Implements error interface.
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

type CreateRobotRequest struct {
	TestClassId string    `json:"testClassId"`
	Scores      string    `json:"scores"`
	Difficulty  string    `json:"difficulty"`
	Type        RobotType `json:"type"`
}

func (CreateRobotRequest) Validate() error {
	return nil
}

type CreateRobotsRequest struct {
	Robots []CreateRobotRequest `json:"robots"`
}

func (CreateRobotsRequest) Validate() error {
	return nil
}

type UpdateRobotRequest struct {
	Scores     string `json:"scores"`
	Difficulty string `json:"difficulty"`
}

func (UpdateRobotRequest) Validate() error {
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

// CustomTime is used to read date values from query parameters
// Implements Convertable and Validable interfaces
type CustomTime time.Time

func (CustomTime) Convert(s string) (CustomTime, error) {
	t, err := time.Parse("2006-01-02", s)
	return CustomTime(t), err
}

func (k CustomTime) AsTime() time.Time {
	return time.Time(k)
}

func (CustomTime) Validate() error {
	return nil
}

type CustomInt64 int64

// CustomInt64 is used to read int value from query parameters.
// Implements Convertable and Validable interfaces
func (CustomInt64) Convert(s string) (CustomInt64, error) {
	a, err := strconv.ParseInt(s, 10, 64)
	return CustomInt64(a), err
}

func (CustomInt64) Validate() error {
	return nil
}

func (k CustomInt64) AsInt64() int64 {
	return int64(k)
}

type CustomString string

// CustomString is a dummy type that implements Convertable and Validable interfaces
func (CustomString) Convert(s string) (CustomString, error) {
	return CustomString(s), nil
}

func (s CustomString) AsString() string {
	return string(s)
}

func (CustomString) Validate() error {
	return nil
}

// CustomInt8 is a dummy type that implements  Convertable and Validable interfaces
type CustomInt8 int8

func (CustomInt8) Convert(s string) (CustomInt8, error) {
	a, err := strconv.ParseInt(s, 10, 8)
	return CustomInt8(a), err
}

func (k CustomInt8) AsInt8() int8 {
	return int8(k)
}

func (CustomInt8) Validate() error {
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

type RobotDto struct {
	ID          int64     `json:"id"`
	TestClassId string    `json:"testClassId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Difficulty  string    `json:"difficulty"`
	Type        RobotType `json:"type"`
	Scores      string    `json:"scores"`
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

func mapToRobotDTO(r *RobotModel) *RobotDto {
	return &RobotDto{
		ID:          r.ID,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		TestClassId: r.TestClassId,
		Difficulty:  r.Difficulty,
		Type:        r.Type,
		Scores:      r.Scores,
	}
}
