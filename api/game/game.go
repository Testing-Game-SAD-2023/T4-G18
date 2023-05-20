package game

import (
	"strconv"
	"time"

	"github.com/alarmfox/game-repository/model"
)

type CreateRequest struct {
	Name    string   `json:"name"`
	Players []string `json:"players"`
}

func (CreateRequest) Validate() error {
	return nil
}

type Game struct {
	ID           int64     `json:"id"`
	CurrentRound int       `json:"currentRound"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	PlayersCount int       `json:"playersCount"`
	Name         string    `json:"name"`
}

type UpdateRequest struct {
	CurrentRound int    `json:"currentRound"`
	Name         string `json:"name"`
}

func (UpdateRequest) Validate() error {
	return nil
}

type Key int64

// Key is used to read int value from query parameters.
// Implements Convertable and Validable interfaces
func (c Key) Parse(s string) (Key, error) {
	a, err := strconv.ParseInt(s, 10, 64)
	return Key(a), err
}

func (Key) Validate() error {
	return nil
}

func (k Key) AsInt64() int64 {
	return int64(k)
}

// Interval is used to read date values from query parameters
// Implements Convertable and Validable interfaces
type Interval time.Time

func (Interval) Parse(s string) (Interval, error) {
	t, err := time.Parse("2006-01-02", s)
	return Interval(t), err
}

func (k Interval) AsTime() time.Time {
	return time.Time(k)
}

func (Interval) Validate() error {
	return nil
}
func fromModel(g *model.Game) Game {
	return Game{
		ID:           g.ID,
		CurrentRound: g.CurrentRound,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
		Name:         g.Name,
		PlayersCount: g.PlayersCount,
	}
}
