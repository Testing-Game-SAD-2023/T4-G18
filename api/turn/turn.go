package turn

import (
	"strconv"
	"time"

	"github.com/alarmfox/game-repository/model"
)

type CreateRequest struct {
	RoundId int64    `json:"roundId"`
	Players []string `json:"players"`
}

func (CreateRequest) Validate() error {
	return nil
}

type UpdateRequest struct {
	Scores   string `json:"scores"`
	IsWinner bool   `json:"isWinner"`
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
func fromModel(t *model.Turn) Turn {
	return Turn{
		ID:        t.ID,
		IsWinner:  t.IsWinner,
		Scores:    t.Scores,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		PlayerID:  t.PlayerID,
	}
}

type Turn struct {
	ID        int64     `json:"id"`
	IsWinner  bool      `json:"isWinner"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	PlayerID  int64     `json:"playerId"`
	Scores    string    `json:"scores"`
}
