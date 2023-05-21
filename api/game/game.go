package game

import (
	"strconv"
	"time"

	"github.com/alarmfox/game-repository/model"
)

type Game struct {
	ID           int64      `json:"id"`
	CurrentRound int        `json:"currentRound"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	StartedAt    *time.Time `json:"startedAt"`
	ClosedAt     *time.Time `json:"closedAt"`
	PlayersCount int        `json:"playersCount"`
	Name         string     `json:"name"`
}

type CreateRequest struct {
	Name      string     `json:"name"`
	Players   []string   `json:"players"`
	StartedAt *time.Time `json:"startedAt,omitempty"`
	ClosedAt  *time.Time `json:"closedAt,omitempty"`
}

func (CreateRequest) Validate() error {
	return nil
}

type UpdateRequest struct {
	CurrentRound int        `json:"currentRound"`
	Name         string     `json:"name"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	ClosedAt     *time.Time `json:"closedAt,omitempty"`
}

func (UpdateRequest) Validate() error {
	return nil
}

type Key int64

func (c Key) Parse(s string) (Key, error) {
	a, err := strconv.ParseInt(s, 10, 64)
	return Key(a), err
}

func (k Key) AsInt64() int64 {
	return int64(k)
}

type Interval time.Time

func (Interval) Parse(s string) (Interval, error) {
	t, err := time.Parse("2006-01-02", s)
	return Interval(t), err
}

func (k Interval) AsTime() time.Time {
	return time.Time(k)
}

func fromModel(g *model.Game) Game {

	return Game{
		ID:           g.ID,
		CurrentRound: g.CurrentRound,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
		Name:         g.Name,
		PlayersCount: g.PlayersCount,
		StartedAt:    g.StartedAt,
		ClosedAt:     g.ClosedAt,
	}

}
