package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"
)

type GameControllerSuite struct {
	suite.Suite
	tServer *httptest.Server
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *GameControllerSuite) SetupSuite() {
	service := NewGameService(&MockedGameRepository{})
	controller := NewGameController(service)

	r := chi.NewMux()
	r.Get("/{id}", makeHTTPHandlerFunc(controller.findByID))
	r.Post("/", makeHTTPHandlerFunc(controller.create))

	suite.tServer = httptest.NewServer(r)
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *GameControllerSuite) TestFindByID() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Arg            string
	}{
		{
			Name:           "T01-GameNotExists",
			ExpectedStatus: http.StatusNotFound,
			Arg:            "12",
		},
		{
			Name:           "T02-GameExists",
			ExpectedStatus: http.StatusOK,
			Arg:            "1",
		},
		{
			Name:           "T03-BadID",
			ExpectedStatus: http.StatusBadRequest,
			Arg:            "aaa",
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			req, err := http.Get(fmt.Sprintf("%s/%s", suite.tServer.URL, tc.Arg))
			suite.NoError(err, "Cannot perform GET request")

			suite.Equal(req.StatusCode, tc.ExpectedStatus)

		})
	}

}
func (suite *GameControllerSuite) TestCreate() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Body           string
	}{
		{
			Name:           "T11-BadJson",
			ExpectedStatus: http.StatusBadRequest,
			Body:           `{"playersCount": 34`,
		},
		{
			Name:           "T12-GameCreated",
			ExpectedStatus: http.StatusCreated,
			Body:           `{"playersCount": 10}`,
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			req, err := http.Post(suite.tServer.URL, "application/json", bytes.NewBufferString(tc.Body))
			suite.NoError(err, "Cannot perform POST request")

			suite.Equal(req.StatusCode, tc.ExpectedStatus)

		})
	}

}

func TestGameControllerSuite(t *testing.T) {
	suite.Run(t, new(GameControllerSuite))
}

func (suite *GameControllerSuite) TearDownSuite() {
	defer suite.tServer.Close()
}

type MockedGameRepository struct {
}

func (gr *MockedGameRepository) Create(r *CreateGameRequest) (*GameModel, error) {
	return &GameModel{}, nil
}

func (gr *MockedGameRepository) FindById(id uint64) (*GameModel, error) {
	if id == 1 {
		return &GameModel{}, nil
	}
	return nil, ErrNotFound
}

func (gr *MockedGameRepository) Delete(id uint64) error {
	if id == 1 {
		return nil
	}
	return ErrNotFound
}
