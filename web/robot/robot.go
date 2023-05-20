package robot

import (
	"fmt"
	"strings"
	"time"

	"github.com/alarmfox/game-repository/model"
	"github.com/alarmfox/game-repository/web"
)

type RobotType int8

const (
	randoop RobotType = iota
	evosuite
)

func (rb RobotType) Parse(s string) (RobotType, error) {
	switch strings.ToLower(s) {
	case "randoop":
		return randoop, nil
	case "evosuite":
		return evosuite, nil
	default:
		return RobotType(2), web.ErrInvalidParam
	}
}
func (rb RobotType) Validate() error {
	switch rb {
	case randoop, evosuite:
		return nil
	default:
		return web.ErrInvalidParam
	}
}

type CreateSingleRequest struct {
	TestClassId string    `json:"testClassId"`
	Scores      string    `json:"scores"`
	Difficulty  string    `json:"difficulty"`
	Type        RobotType `json:"type"`
}

func (r CreateSingleRequest) Validate() error {
	switch r.Type {
	case randoop, evosuite:
		return nil
	default:
		return fmt.Errorf("%w: unsupported robot type %q", web.ErrInvalidParam, r.Type)
	}
}

type CreateRequest struct {
	Robots []CreateSingleRequest `json:"robots"`
}

func (robots CreateRequest) Validate() error {
	for _, robot := range robots.Robots {
		if err := robot.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type UpdateRequest struct {
	Scores     string `json:"scores"`
	Difficulty string `json:"difficulty"`
}

func (UpdateRequest) Validate() error {
	return nil
}

type CustomString string

// CustomString is a dummy type that implements Convertable and Validable interfaces
func (CustomString) Parse(s string) (CustomString, error) {
	return CustomString(s), nil
}

func (s CustomString) AsString() string {
	return string(s)
}

func (CustomString) Validate() error {
	return nil
}

type Robot struct {
	ID          int64     `json:"id"`
	TestClassId string    `json:"testClassId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Difficulty  string    `json:"difficulty"`
	Type        RobotType `json:"type"`
	Scores      string    `json:"scores"`
}

func fromModel(r *model.Robot) *Robot {
	return &Robot{
		ID:          r.ID,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		TestClassId: r.TestClassId,
		Difficulty:  r.Difficulty,
		Type:        RobotType(r.Type),
		Scores:      r.Scores,
	}
}
