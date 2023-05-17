package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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
		Return(nil, ErrNotFound)

	controller := NewGameController(gr)

	r := chi.NewMux()
	r.Get("/{id}", makeHTTPHandlerFunc(controller.findByID))
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
			suite.NoError(err)

			suite.Equal(tc.ExpectedStatus, req.StatusCode, err)

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
			Name:           "T04-BadJson",
			ExpectedStatus: http.StatusBadRequest,
			Body:           `{"playersCount": 34`,
		},
		{
			Name:           "T05-GameCreated",
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
			suite.Equal(tc.ExpectedStatus, req.StatusCode, err)

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
			Name:           "T06-BadId",
			ExpectedStatus: http.StatusBadRequest,
			Id:             `26a4`,
		},
		{
			Name:           "T07-GameDeleted",
			ExpectedStatus: http.StatusNoContent,
			Id:             `1`,
		},
		{
			Name:           "T08-GameNotFound",
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

			suite.Equal(tc.ExpectedStatus, res.StatusCode, err)

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
			Name:           "T09-BadId",
			ExpectedStatus: http.StatusBadRequest,
			Body:           ` {"currentRound": 10, "name: "test"}`,
			Id:             `1a`,
		},
		{
			Name:           "T010-GameUpdated",
			ExpectedStatus: http.StatusOK,
			Body:           `{"currentRound": 10, "name": "test"}`,
			Id:             `1`,
		},
		{
			Name:           "T011-InvalidJSON",
			ExpectedStatus: http.StatusBadRequest,
			Body:           ` {"currentRound": 10, "name: "test"}`,
			Id:             `1`,
		},
		{
			Name:           "T012-GameNotFound",
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
			suite.Equal(tc.ExpectedStatus, res.StatusCode, err)

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
	return nil, 0, fmt.Errorf("not implemented")
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
		Return(ErrNotFound)

	suite.tmpDir = os.TempDir()
	controller := NewTurnController(tr)

	r := chi.NewMux()

	r.Get("/{id}/files", makeHTTPHandlerFunc(controller.download))
	r.Put("/{id}/files", makeHTTPHandlerFunc(controller.upload))

	suite.tServer = httptest.NewServer(r)
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
			suite.Equal(tc.ExpectedStatus, res.StatusCode, err)

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
	return nil, fmt.Errorf("not implemented")
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
	return fmt.Errorf("not implemented")
}

func (m *MockedTurnRepository) FindByRound(id int64) ([]TurnModel, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockedTurnRepository) Update(id int64, request *UpdateTurnRequest) (*TurnModel, error) {
	return nil, fmt.Errorf("not implemented")
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
