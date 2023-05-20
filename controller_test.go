package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type GameControllerSuite struct {
	suite.Suite
	tServer *httptest.Server
}

func (suite *GameControllerSuite) SetupSuite() {
	gr := new(MockedGameRepository)
	gr.
		On("Create", &CreateGameRequest{Name: ""}).
		Return(&GameModel{ID: 1}, nil).
		On("FindById", int64(1)).
		Return(&GameModel{ID: 1}, nil).
		On("FindById",
			mock.MatchedBy(func(id int64) bool { return id != 1 })).
		Return(nil, ErrNotFound).
		On("Delete", int64(1)).
		Return(nil).
		On("Delete",
			mock.MatchedBy(func(id int64) bool { return id != 1 })).
		Return(ErrNotFound).
		On("Update", int64(1),
			&UpdateGameRequest{Name: "test", CurrentRound: 10}).
		Return(&GameModel{}, nil).
		On("Update",
			mock.MatchedBy(func(id int64) bool { return id != 1 }),
			&UpdateGameRequest{Name: "test", CurrentRound: 10}).
		Return(nil, ErrNotFound).
		On("FindByInterval", mock.Anything, mock.Anything).
		Return([]GameModel{}, int(64), nil)

	controller := NewGameController(gr)

	r := chi.NewMux()
	r.Get("/{id}", makeHTTPHandlerFunc(controller.findByID))
	r.Get("/", makeHTTPHandlerFunc(controller.list))
	r.Post("/", makeHTTPHandlerFunc(controller.create))
	r.Delete("/{id}", makeHTTPHandlerFunc(controller.delete))
	r.Put("/{id}", makeHTTPHandlerFunc(controller.update))

	suite.tServer = httptest.NewServer(r)
}

func (suite *GameControllerSuite) TestFindByID() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Arg            string
	}{
		{
			Name:           "T00-01-GameNotExists",
			ExpectedStatus: http.StatusNotFound,
			Arg:            "12",
		},
		{
			Name:           "T00-02-GameExists",
			ExpectedStatus: http.StatusOK,
			Arg:            "1",
		},
		{
			Name:           "T00-03-BadID",
			ExpectedStatus: http.StatusBadRequest,
			Arg:            "aaa",
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			req, err := http.Get(fmt.Sprintf("%s/%s", suite.tServer.URL, tc.Arg))
			suite.NoError(err)

			suite.Equal(tc.ExpectedStatus, req.StatusCode, tc.Name)

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
			Name:           "T00-04-BadJson",
			ExpectedStatus: http.StatusBadRequest,
			Body:           `{"playersCount": 34`,
		},
		{
			Name:           "T00-05-GameCreated",
			ExpectedStatus: http.StatusCreated,
			Body:           `{"playersCount": 10}`,
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			req, err := http.Post(suite.tServer.URL,
				"application/json",
				bytes.NewBufferString(tc.Body))

			suite.NoError(err)
			suite.Equal(tc.ExpectedStatus, req.StatusCode, tc.Name)

		})
	}

}

func (suite *GameControllerSuite) TestDelete() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Id             string
	}{
		{
			Name:           "T00-06-BadId",
			ExpectedStatus: http.StatusBadRequest,
			Id:             `26a4`,
		},
		{
			Name:           "T00-07-GameDeleted",
			ExpectedStatus: http.StatusNoContent,
			Id:             `1`,
		},
		{
			Name:           "T00-08-GameNotFound",
			ExpectedStatus: http.StatusNotFound,
			Id:             `14`,
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s", suite.tServer.URL, tc.Id)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			suite.NoError(err)

			res, err := http.DefaultClient.Do(req)
			suite.NoError(err)

			suite.Equal(tc.ExpectedStatus, res.StatusCode, tc.Name)

		})
	}

}
func (suite *GameControllerSuite) TestUpdate() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Body           string
		Id             string
	}{
		{
			Name:           "T00-09-BadId",
			ExpectedStatus: http.StatusBadRequest,
			Body:           ` {"currentRound": 10, "name: "test"}`,
			Id:             `1a`,
		},
		{
			Name:           "T00-10-GameUpdated",
			ExpectedStatus: http.StatusOK,
			Body:           `{"currentRound": 10, "name": "test"}`,
			Id:             `1`,
		},
		{
			Name:           "T00-11-InvalidJSON",
			ExpectedStatus: http.StatusBadRequest,
			Body:           ` {"currentRound": 10, "name: "test"}`,
			Id:             `1`,
		},
		{
			Name:           "T00-12-GameNotFound",
			ExpectedStatus: http.StatusNotFound,
			Body:           ` {"currentRound": 10, "name": "test"}`,
			Id:             `11`,
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s", suite.tServer.URL, tc.Id)
			req, err := http.NewRequest(http.MethodPut,
				url,
				bytes.NewBufferString(tc.Body))

			suite.NoError(err)
			res, err := http.DefaultClient.Do(req)
			suite.NoError(err)
			suite.Equal(tc.ExpectedStatus, res.StatusCode, tc.Name)

		})
	}

}
func (suite *GameControllerSuite) TestList() {

	type input struct {
		Page      string
		PageSize  string
		StartDate string
		EndDate   string
	}

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Input          input
	}{
		{
			Name:           "Valid input",
			ExpectedStatus: http.StatusOK,
			Input: input{
				Page:      "1",
				PageSize:  "10",
				StartDate: "2023-01-01",
				EndDate:   "2023-01-31",
			},
		},
		{
			Name:           "Invalid page",
			ExpectedStatus: http.StatusBadRequest,
			Input: input{
				Page:      "invalid",
				PageSize:  "10",
				StartDate: "2023-01-01",
				EndDate:   "2023-01-31",
			},
		},
		{
			Name:           "Invalid pageSize",
			ExpectedStatus: http.StatusBadRequest,
			Input: input{
				Page:      "1",
				PageSize:  "invalid",
				StartDate: "2023-01-01",
				EndDate:   "2023-01-31",
			},
		},
		{
			Name:           "Invalid startDate",
			ExpectedStatus: http.StatusBadRequest,
			Input: input{
				Page:      "1",
				PageSize:  "10",
				StartDate: "invalid",
				EndDate:   "2023-01-31",
			},
		},
		{
			Name:           "Invalid endDate",
			ExpectedStatus: http.StatusBadRequest,
			Input: input{
				Page:      "1",
				PageSize:  "10",
				StartDate: "2023-01-01",
				EndDate:   "invalid",
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			q := url.Values{}
			q.Set("page", tc.Input.Page)
			q.Set("pageSize", tc.Input.PageSize)
			q.Set("startDate", tc.Input.StartDate)
			q.Set("endDate", tc.Input.EndDate)
			url := fmt.Sprintf("%s?%s", suite.tServer.URL, q.Encode())
			res, err := http.Get(url)
			suite.NoError(err)
			suite.Equal(tc.ExpectedStatus, res.StatusCode, tc.Name)

			defer res.Body.Close()
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
	mock.Mock
}

func (gr *MockedGameRepository) Create(r *CreateGameRequest) (*GameModel, error) {
	args := gr.Called(r)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.(*GameModel), args.Error(1)
}

func (gr *MockedGameRepository) FindById(id int64) (*GameModel, error) {
	args := gr.Called(id)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.(*GameModel), args.Error(1)

}

func (gr *MockedGameRepository) Delete(id int64) error {
	args := gr.Called(id)
	return args.Error(0)
}

func (gr *MockedGameRepository) Update(id int64, ur *UpdateGameRequest) (*GameModel, error) {
	args := gr.Called(id, ur)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.(*GameModel), args.Error(1)
}

func (gr *MockedGameRepository) FindByInterval(i *IntervalParams, p *PaginationParams) ([]GameModel, int64, error) {
	args := gr.Called(i, p)
	v := args.Get(0)

	if v == nil {
		return nil, int64(args.Int(1)), args.Error(2)
	}
	return v.([]GameModel), int64(args.Int(1)), args.Error(2)

}

type RoundControllerSuite struct {
	suite.Suite
	tServer *httptest.Server
}

func (suite *RoundControllerSuite) SetupSuite() {
	rr := new(MockedRoundRepository)
	rr.
		On("Create", &CreateRoundRequest{GameId: 1, TestClassId: "a.java", Order: 1}).
		Return(&RoundModel{ID: 1}, nil).
		On("Create",
			mock.MatchedBy(func(r *CreateRoundRequest) bool { return r.GameId != 1 })).
		Return(nil, ErrNotFound).
		On("FindById", int64(1)).
		Return(&RoundModel{ID: 1}, nil).
		On("FindById",
			mock.MatchedBy(func(id int64) bool { return id != 1 })).
		Return(nil, ErrNotFound).
		On("Delete", int64(1)).
		Return(nil).
		On("Delete",
			mock.MatchedBy(func(id int64) bool { return id != 1 })).
		Return(ErrNotFound).
		On("Update", int64(1),
			&UpdateRoundRequest{Order: 2}).
		Return(&RoundModel{}, nil).
		On("Update",
			mock.MatchedBy(func(id int64) bool { return id != 1 }),
			&UpdateRoundRequest{Order: 2}).
		Return(nil, ErrNotFound).
		On("FindByGame", int64(1)).
		Return([]RoundModel{}, nil).
		On("FindByGame", mock.MatchedBy(func(id int64) bool { return id != 1 })).
		Return(nil, ErrNotFound)

	controller := NewRoundController(rr)

	r := chi.NewMux()
	r.Get("/{id}", makeHTTPHandlerFunc(controller.findByID))
	r.Get("/", makeHTTPHandlerFunc(controller.list))
	r.Post("/", makeHTTPHandlerFunc(controller.create))
	r.Delete("/{id}", makeHTTPHandlerFunc(controller.delete))
	r.Put("/{id}", makeHTTPHandlerFunc(controller.update))

	suite.tServer = httptest.NewServer(r)
}

func (suite *RoundControllerSuite) TestFindByID() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Arg            string
	}{
		{
			Name:           "T01-01-RoundNotExists",
			ExpectedStatus: http.StatusNotFound,
			Arg:            "12",
		},
		{
			Name:           "T01-02-RoundExists",
			ExpectedStatus: http.StatusOK,
			Arg:            "1",
		},
		{
			Name:           "T01-03-BadID",
			ExpectedStatus: http.StatusBadRequest,
			Arg:            "aaa",
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			req, err := http.Get(fmt.Sprintf("%s/%s", suite.tServer.URL, tc.Arg))
			suite.NoError(err)

			suite.Equal(tc.ExpectedStatus, req.StatusCode, tc.Name)

		})
	}
}

func (suite *RoundControllerSuite) TestCreate() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Body           string
	}{
		{
			Name:           "T01-04-BadJson",
			ExpectedStatus: http.StatusBadRequest,
			Body:           `{"playersCount": 34`,
		},
		{
			Name:           "T01-05-RoundCreated",
			ExpectedStatus: http.StatusCreated,
			Body:           `{"gameId": 1, "testClassId": "a.java", "order": 1}`,
		},
		{
			Name:           "T01-06-GameNotExists",
			ExpectedStatus: http.StatusNotFound,
			Body:           `{"gameId": 2}`,
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			req, err := http.Post(suite.tServer.URL,
				"application/json",
				bytes.NewBufferString(tc.Body))

			suite.NoError(err)
			suite.Equal(tc.ExpectedStatus, req.StatusCode, tc.Name)

		})
	}

}

func (suite *RoundControllerSuite) TestDelete() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Id             string
	}{
		{
			Name:           "T01-07-BadId",
			ExpectedStatus: http.StatusBadRequest,
			Id:             `26a4`,
		},
		{
			Name:           "T01-08-RoundDeleted",
			ExpectedStatus: http.StatusNoContent,
			Id:             `1`,
		},
		{
			Name:           "T01-09-RoundNotFound",
			ExpectedStatus: http.StatusNotFound,
			Id:             `14`,
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s", suite.tServer.URL, tc.Id)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			suite.NoError(err)

			res, err := http.DefaultClient.Do(req)
			suite.NoError(err)

			suite.Equal(tc.ExpectedStatus, res.StatusCode, tc.Name)

		})
	}

}
func (suite *RoundControllerSuite) TestUpdate() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Body           string
		Id             string
	}{
		{
			Name:           "T01-10-BadId",
			ExpectedStatus: http.StatusBadRequest,
			Body:           ` {"order": 1}`,
			Id:             `1a`,
		},
		{
			Name:           "T01-11-RoundUpdated",
			ExpectedStatus: http.StatusOK,
			Body:           `{"order": 2}`,
			Id:             `1`,
		},
		{
			Name:           "T01-12-InvalidJSON",
			ExpectedStatus: http.StatusBadRequest,
			Body:           ` {"order": 1t"}`,
			Id:             `1`,
		},
		{
			Name:           "T01-13-RoundNotFound",
			ExpectedStatus: http.StatusNotFound,
			Body:           ` {"order": 2}`,
			Id:             `11`,
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s", suite.tServer.URL, tc.Id)
			req, err := http.NewRequest(http.MethodPut,
				url,
				bytes.NewBufferString(tc.Body))

			suite.NoError(err)
			res, err := http.DefaultClient.Do(req)
			suite.NoError(err)
			suite.Equal(tc.ExpectedStatus, res.StatusCode, tc.Name)

		})
	}

}
func (suite *RoundControllerSuite) TestList() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Input          string
	}{
		{
			Name:           "T01-14-OkList",
			ExpectedStatus: http.StatusOK,
			Input:          "1",
		},
		{
			Name:           "T01-15-InvalidId",
			ExpectedStatus: http.StatusBadRequest,
			Input:          "invalid",
		},
		{
			Name:           "T01-16-RoundNotFound",
			ExpectedStatus: http.StatusNotFound,
			Input:          "2",
		},
	}

	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			q := url.Values{}
			q.Add("gameId", tc.Input)
			url := fmt.Sprintf("%s?%s", suite.tServer.URL, q.Encode())
			res, err := http.Get(url)
			suite.NoError(err)
			suite.Equal(tc.ExpectedStatus, res.StatusCode, tc.Name)

			defer res.Body.Close()
		})
	}

}

func TestRoundControllerSuite(t *testing.T) {
	suite.Run(t, new(RoundControllerSuite))
}

func (suite *RoundControllerSuite) TearDownSuite() {
	defer suite.tServer.Close()
}

type MockedRoundRepository struct {
	mock.Mock
}

func (gr *MockedRoundRepository) Create(r *CreateRoundRequest) (*RoundModel, error) {
	args := gr.Called(r)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.(*RoundModel), args.Error(1)
}

func (gr *MockedRoundRepository) FindById(id int64) (*RoundModel, error) {
	args := gr.Called(id)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.(*RoundModel), args.Error(1)

}

func (gr *MockedRoundRepository) Delete(id int64) error {
	args := gr.Called(id)
	return args.Error(0)
}

func (gr *MockedRoundRepository) Update(id int64, ur *UpdateRoundRequest) (*RoundModel, error) {
	args := gr.Called(id, ur)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.(*RoundModel), args.Error(1)
}

func (gr *MockedRoundRepository) FindByGame(id int64) ([]RoundModel, error) {
	args := gr.Called(id)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.([]RoundModel), args.Error(1)

}

type TurnControllerSuite struct {
	suite.Suite
	tServer *httptest.Server
	tmpDir  string
}

func (suite *TurnControllerSuite) SetupSuite() {
	tr := new(MockedTurnRepository)
	f, err := os.CreateTemp(suite.tmpDir, "")
	suite.NoError(err)
	_, err = f.Write([]byte("hello"))
	suite.NoError(err)

	tr.
		On("GetFile",
			int64(1)).
		Return(f.Name(), f, nil).
		On("GetFile",
			mock.MatchedBy(func(id int64) bool { return id != 1 })).
		Return("", nil, ErrNotFound).
		On("SaveFile",
			int64(1),
			mock.Anything).
		Return(nil).
		On("SaveFile",
			mock.MatchedBy(func(id int64) bool { return id != 1 }),
			mock.Anything).
		Return(ErrNotFound).
		On("CreateBulk", &CreateTurnsRequest{RoundId: 1}).
		Return([]TurnModel{}, nil).
		On("CreateBulk", mock.MatchedBy(func(r *CreateTurnsRequest) bool { return r.RoundId != 1 })).
		Return(nil, ErrNotFound).
		On("FindById", int64(1)).
		Return(&TurnModel{ID: 1}, nil).
		On("FindById",
			mock.MatchedBy(func(id int64) bool { return id != 1 })).
		Return(nil, ErrNotFound).
		On("Delete", int64(1)).
		Return(nil).
		On("Delete",
			mock.MatchedBy(func(id int64) bool { return id != 1 })).
		Return(ErrNotFound).
		On("Update", int64(1), &UpdateTurnRequest{IsWinner: true, Scores: "a"}).
		Return(&TurnModel{}, nil).
		On("Update", mock.MatchedBy(func(id int64) bool { return id != 1 }),
			&UpdateTurnRequest{IsWinner: true, Scores: "a"}).
		Return(nil, ErrNotFound).
		On("FindByRound", int64(1)).
		Return([]TurnModel{}, nil).
		On("FindByRound", mock.MatchedBy(func(id int64) bool { return id != 1 })).
		Return(nil, ErrNotFound)

	suite.tmpDir = os.TempDir()
	controller := NewTurnController(tr)

	r := chi.NewMux()

	r.Get("/{id}/files", makeHTTPHandlerFunc(controller.download))
	r.Put("/{id}/files", makeHTTPHandlerFunc(controller.upload))
	r.Post("/", makeHTTPHandlerFunc(controller.create))
	r.Get("/", makeHTTPHandlerFunc(controller.list))
	r.Put("/{id}", makeHTTPHandlerFunc(controller.update))
	r.Delete("/{id}", makeHTTPHandlerFunc(controller.delete))
	r.Get("/{id}", makeHTTPHandlerFunc(controller.findByID))

	suite.tServer = httptest.NewServer(r)
}

func (suite *TurnControllerSuite) TestFindByID() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Arg            string
	}{
		{
			Name:           "T02-01-TurnNotExists",
			ExpectedStatus: http.StatusNotFound,
			Arg:            "12",
		},
		{
			Name:           "T02-02-TurnExists",
			ExpectedStatus: http.StatusOK,
			Arg:            "1",
		},
		{
			Name:           "T02-03-BadID",
			ExpectedStatus: http.StatusBadRequest,
			Arg:            "aaa",
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			req, err := http.Get(fmt.Sprintf("%s/%s", suite.tServer.URL, tc.Arg))
			suite.NoError(err)

			suite.Equal(tc.ExpectedStatus, req.StatusCode, tc.Name)

		})
	}
}

func (suite *TurnControllerSuite) TestCreate() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Body           string
	}{
		{
			Name:           "T02-04-BadJson",
			ExpectedStatus: http.StatusBadRequest,
			Body:           `{"roundId": 34`,
		},
		{
			Name:           "T02-05-TurnCreated",
			ExpectedStatus: http.StatusCreated,
			Body:           `{"roundId": 1, "playerId": 1}`,
		},
		{
			Name:           "T02-06-RoundNotExists",
			ExpectedStatus: http.StatusNotFound,
			Body:           `{"turnId": 2}`,
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			req, err := http.Post(suite.tServer.URL,
				"application/json",
				bytes.NewBufferString(tc.Body))

			suite.NoError(err)
			suite.Equal(tc.ExpectedStatus, req.StatusCode, tc.Name)

		})
	}

}

func (suite *TurnControllerSuite) TestDelete() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Id             string
	}{
		{
			Name:           "T02-07-BadId",
			ExpectedStatus: http.StatusBadRequest,
			Id:             `26a4`,
		},
		{
			Name:           "T02-08-TurnDeleted",
			ExpectedStatus: http.StatusNoContent,
			Id:             `1`,
		},
		{
			Name:           "T02-09-TurnNotFound",
			ExpectedStatus: http.StatusNotFound,
			Id:             `14`,
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s", suite.tServer.URL, tc.Id)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			suite.NoError(err)

			res, err := http.DefaultClient.Do(req)
			suite.NoError(err)

			suite.Equal(tc.ExpectedStatus, res.StatusCode, tc.Name)

		})
	}

}
func (suite *TurnControllerSuite) TestUpdate() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Body           string
		Id             string
	}{
		{
			Name:           "T02-10-BadId",
			ExpectedStatus: http.StatusBadRequest,
			Body:           `{"scores": "a", "isWinner": true}`,
			Id:             `1a`,
		},
		{
			Name:           "T02-11-TurnUpdated",
			ExpectedStatus: http.StatusOK,
			Body:           `{"scores": "a", "isWinner": true}`,
			Id:             `1`,
		},
		{
			Name:           "T02-12-InvalidJSON",
			ExpectedStatus: http.StatusBadRequest,
			Body:           `{"order"}`,
			Id:             `1`,
		},
		{
			Name:           "T02-13-TurnNotFound",
			ExpectedStatus: http.StatusNotFound,
			Body:           `{"scores": "a", "isWinner": true}`,
			Id:             `11`,
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			url := fmt.Sprintf("%s/%s", suite.tServer.URL, tc.Id)
			req, err := http.NewRequest(http.MethodPut,
				url,
				bytes.NewBufferString(tc.Body))

			suite.NoError(err)
			res, err := http.DefaultClient.Do(req)
			suite.NoError(err)
			suite.Equal(tc.ExpectedStatus, res.StatusCode, tc.Name)

		})
	}

}
func (suite *TurnControllerSuite) TestList() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		Input          string
	}{
		{
			Name:           "T02-14-TurnsFound",
			ExpectedStatus: http.StatusOK,
			Input:          "1",
		},
		{
			Name:           "T02-15-ErrBadId",
			ExpectedStatus: http.StatusBadRequest,
			Input:          "invalid",
		},
		{
			Name:           "T02-16-RoundNotFound",
			ExpectedStatus: http.StatusNotFound,
			Input:          "2",
		},
	}

	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			q := url.Values{}
			q.Add("roundId", tc.Input)
			url := fmt.Sprintf("%s?%s", suite.tServer.URL, q.Encode())
			res, err := http.Get(url)
			suite.NoError(err)
			suite.Equal(tc.ExpectedStatus, res.StatusCode, tc.Name)

			defer res.Body.Close()
		})
	}

}

func (suite *TurnControllerSuite) TestUpload() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		TurnID         string
		Body           io.Reader
	}{
		{
			Name:           "T21-BadZip",
			ExpectedStatus: http.StatusOK,
			TurnID:         "1",
			Body:           bytes.NewBufferString("not a zip"),
		},
		{
			Name:           "T22-ZipSaved",
			ExpectedStatus: http.StatusOK,
			TurnID:         "1",
			Body:           generateValidZipContent(suite.T(), []byte("hello")),
		},
		{
			Name:           "T23-TurnNotFound",
			ExpectedStatus: http.StatusNotFound,
			TurnID:         "12",
			Body:           generateValidZipContent(suite.T(), []byte("hello")),
		},
		{
			Name:           "T24-BadTurnID",
			ExpectedStatus: http.StatusBadRequest,
			TurnID:         "a12",
			Body:           generateValidZipContent(suite.T(), []byte("hell")),
		},
	}

	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {

			req, err := http.NewRequest(http.MethodPut,
				fmt.Sprintf("%s/%s/files",
					suite.tServer.URL,
					tc.TurnID),
				tc.Body)

			suite.NoError(err)

			res, err := http.DefaultClient.Do(req)
			suite.NoError(err)
			suite.Equal(tc.ExpectedStatus, res.StatusCode, tc.Name)

		})
	}
}

func (suite *TurnControllerSuite) TestDownload() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		TurnID         string
	}{
		{
			Name:           "T35-DownloadOK",
			ExpectedStatus: http.StatusOK,
			TurnID:         "1",
		},
		{
			Name:           "T36-TurnNotFound",
			ExpectedStatus: http.StatusNotFound,
			TurnID:         "21",
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {

			res, err := http.Get(fmt.Sprintf("%s/%s/files", suite.tServer.URL, tc.TurnID))
			suite.NoError(err)
			defer res.Body.Close()

			_, err = io.Copy(io.Discard, res.Body)

			suite.NoError(err)
			suite.Equal(tc.ExpectedStatus, res.StatusCode, tc.Name)

		})
	}
}

func (suite *TurnControllerSuite) TearDownSuite() {
	defer os.RemoveAll(suite.tmpDir)
	defer suite.tServer.Close()
}
func TestTurnControllerSuite(t *testing.T) {
	suite.Run(t, new(TurnControllerSuite))
}

type MockedTurnRepository struct {
	mock.Mock
}

func (m *MockedTurnRepository) CreateBulk(request *CreateTurnsRequest) ([]TurnModel, error) {
	args := m.Called(request)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.([]TurnModel), args.Error(1)
}

func (m *MockedTurnRepository) FindById(id int64) (*TurnModel, error) {
	args := m.Called(id)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.(*TurnModel), args.Error(1)
}

func (m *MockedTurnRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockedTurnRepository) FindByRound(id int64) ([]TurnModel, error) {
	args := m.Called(id)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.([]TurnModel), args.Error(1)
}

func (m *MockedTurnRepository) Update(id int64, request *UpdateTurnRequest) (*TurnModel, error) {
	args := m.Called(id, request)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.(*TurnModel), args.Error(1)
}

func (m *MockedTurnRepository) SaveFile(id int64, r io.Reader) error {
	args := m.Called(id, r)
	return args.Error(0)
}

func (m *MockedTurnRepository) GetFile(id int64) (string, *os.File, error) {
	args := m.Called(id)
	v := args.Get(1)

	if v == nil {
		return "", nil, args.Error(2)
	}
	return args.String(0), v.(*os.File), args.Error(2)
}

type RobotControllerSuite struct {
	suite.Suite
	tServer *httptest.Server
}

func (suite *RobotControllerSuite) SetupSuite() {
	tr := new(MockedRobotRepository)

	tr.
		On("CreateBulk", mock.Anything).
		Return(1, nil).
		On("FindByFilter", mock.Anything, mock.Anything, mock.Anything).
		Return(&RobotModel{ID: 1}, nil).
		On("DeleteByTestClass", "a.java").
		Return(nil).
		On("DeleteByTestClass",
			mock.MatchedBy(func(id string) bool { return id != "a.java" })).
		Return(ErrNotFound)

	controller := NewRobotController(tr)

	r := chi.NewMux()

	r.Post("/", makeHTTPHandlerFunc(controller.createBulk))
	r.Get("/", makeHTTPHandlerFunc(controller.findByFilter))
	r.Delete("/", makeHTTPHandlerFunc(controller.delete))

	suite.tServer = httptest.NewServer(r)
}
func (suite *RobotControllerSuite) TearDownSuite() {
	defer suite.tServer.Close()
}
func TestRobotControllerSuite(t *testing.T) {
	suite.Run(t, new(RobotControllerSuite))
}

func (suite *RobotControllerSuite) TestFindByFilter() {
	type input struct {
		TestClassId string
		Difficulty  string
		RobotType   string
	}
	tcs := []struct {
		Name           string
		ExpectedStatus int
		Input          input
	}{
		{
			Name:           "T04-01-ValidInput",
			ExpectedStatus: http.StatusOK,
			Input: input{
				TestClassId: "TestRobot.java",
				Difficulty:  "easy",
				RobotType:   "0",
			},
		},
		{
			Name:           "T04-02-ValidInput",
			ExpectedStatus: http.StatusOK,
			Input: input{
				TestClassId: "SomeOtherRobot.java",
				Difficulty:  "medium",
				RobotType:   "1",
			},
		},
		{
			Name:           "T04-03-InvalidInput",
			ExpectedStatus: http.StatusBadRequest,
			Input: input{
				TestClassId: "InvalidRobot.java",
				Difficulty:  "hard",
				RobotType:   "2",
			},
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			q := url.Values{}
			q.Set("testClassId", tc.Input.TestClassId)
			q.Set("difficulty", tc.Input.Difficulty)
			q.Set("type", tc.Input.RobotType)

			req, err := http.Get(fmt.Sprintf("%s?%s", suite.tServer.URL, q.Encode()))
			suite.NoError(err)
			suite.Equal(tc.ExpectedStatus, req.StatusCode, tc.Name)

		})
	}
}
func (suite *RobotControllerSuite) TestCreate() {
	tcs := []struct {
		Name           string
		ExpectedStatus int
		Body           string
	}{
		{
			Name:           "T04-04-ValidInput",
			ExpectedStatus: http.StatusCreated,
			Body:           `{"robots": [{"testClassId": "a.java", "scores": "some scores", "difficulty": "some difficulty", "type": 0}]}`,
		},
		{
			Name:           "T04-05-InvalidRobotType",
			ExpectedStatus: http.StatusBadRequest,
			Body:           `{"robots": [{"testClassId": "a.java", "scores": "some scores", "difficulty": "some difficulty", "type": 2}]}`,
		},
		{
			Name:           "T04-06-MissingField",
			ExpectedStatus: http.StatusCreated,
			Body:           `{"robots": [{"testClassId": "a.java", "scores": "some scores", "difficulty": "some difficulty"}]}`,
		},
		{
			Name:           "T04-07-BadlyFormattedJSON",
			ExpectedStatus: http.StatusBadRequest,
			Body:           `{"robots: [{"testClassId": "a.java", "scores": "some scores", "difficulty": "some difficulty", "type": 1}]}`,
		},
	}

	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			req, err := http.Post(suite.tServer.URL,
				"application/json",
				bytes.NewBufferString(tc.Body))

			suite.NoError(err)
			suite.Equal(tc.ExpectedStatus, req.StatusCode, tc.Name)

		})
	}
}

func (suite *RobotControllerSuite) TestDelete() {

	tcs := []struct {
		Name           string
		ExpectedStatus int
		TestClassId    string
	}{
		{
			Name:           "T04-08-Ok",
			ExpectedStatus: http.StatusNoContent,
			TestClassId:    `a.java`,
		},
		{
			Name:           "T04-09-NotFound",
			ExpectedStatus: http.StatusNotFound,
			TestClassId:    `b.java`,
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {
			p := url.Values{}
			p.Add("testClassId", tc.TestClassId)
			url := fmt.Sprintf("%s?%s", suite.tServer.URL, p.Encode())
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			suite.NoError(err)

			res, err := http.DefaultClient.Do(req)
			suite.NoError(err)

			suite.Equal(tc.ExpectedStatus, res.StatusCode, tc.Name)

		})
	}

}

type MockedRobotRepository struct {
	mock.Mock
}

func (m *MockedRobotRepository) CreateBulk(request *CreateRobotsRequest) (int, error) {
	args := m.Called(request)

	return args.Int(0), args.Error(1)
}

func (m *MockedRobotRepository) FindByFilter(testClassId string, difficulty string, t RobotType) (*RobotModel, error) {
	args := m.Called(testClassId, difficulty, t)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.(*RobotModel), args.Error(1)

}
func (m *MockedRobotRepository) DeleteByTestClass(testClassId string) error {
	args := m.Called(testClassId)
	return args.Error(0)
}
