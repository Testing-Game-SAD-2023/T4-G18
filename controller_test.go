package main

import (
	"archive/zip"
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
		On("Create", &CreateGameRequest{PlayersCount: 10}).Return(&GameModel{ID: 1}, nil).
		On("FindById", int64(1)).Return(&GameModel{ID: 1}, nil).
		On("FindById", mock.MatchedBy(func(id int64) bool { return id != 1 })).Return(nil, ErrNotFound).
		On("Delete", int64(1)).Return(nil).
		On("Delete", mock.MatchedBy(func(id int64) bool { return id != 1 })).Return(ErrNotFound).
		On("Update", int64(1), &UpdateGameRequest{Name: "test", CurrentRound: 10}).Return(&GameModel{}, nil).
		On("Update", mock.MatchedBy(func(id int64) bool { return id != 1 }), &UpdateGameRequest{Name: "test", CurrentRound: 10}).Return(nil, ErrNotFound)

	service := NewGameService(gr)
	controller := NewGameController(service)

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
			suite.NoError(err, "Cannot perform GET request")

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
			req, err := http.Post(suite.tServer.URL, "application/json", bytes.NewBufferString(tc.Body))
			suite.NoError(err, "Cannot perform POST request")

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
			req, err := http.NewRequest(http.MethodPut, url, bytes.NewBufferString(tc.Body))
			suite.NoError(err)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			suite.NoError(err)

			suite.Equal(tc.ExpectedStatus, res.StatusCode, tc.Name)

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

func (gr *MockedGameRepository) FindByRound(id int64) (*GameModel, error) {
	args := gr.Called(id)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.(*GameModel), args.Error(1)

}

type TurnControllerSuite struct {
	suite.Suite
	tServer *httptest.Server
	tmpDir  string
}

func (suite *TurnControllerSuite) SetupSuite() {
	tr := new(MockedTurnRepository)
	mr := new(MockedMetadataRepository)
	gr := new(MockedGameRepository)
	f, err := os.CreateTemp(suite.tmpDir, "")
	suite.NoError(err)
	_, err = f.Write([]byte("hello"))
	suite.NoError(err)
	defer f.Close()

	tr.
		On("FindById", int64(1)).Return(&TurnModel{ID: 1}, nil).
		On("FindById", mock.MatchedBy(func(id int64) bool { return id != 1 })).Return(nil, ErrNotFound)

	mr.
		On("Upsert", int64(1)).Return(nil).
		On("Upsert", mock.MatchedBy(func(id int64) bool { return id != 1 })).Return(ErrNotFound).
		On("FindByTurn", int64(1)).Return(&MetadataModel{Path: f.Name()}, nil).
		On("FindByTurn", mock.MatchedBy(func(id int64) bool { return id != 1 })).Return(nil, ErrNotFound)

	gr.
		On("FindByRound", int64(1)).Return(&GameModel{ID: 1}, nil).
		On("FindByRound", mock.MatchedBy(func(id int64) bool { return id != 1 })).Return(&GameModel{ID: 1}, nil)

	suite.tmpDir = os.TempDir()
	service := NewTurnService(tr, mr, gr, suite.tmpDir)
	controller := NewTurnController(service, 512)

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
		BodyFn         func() *bytes.Buffer
	}{
		{
			Name:           "T21-BadZip",
			ExpectedStatus: http.StatusUnprocessableEntity,
			TurnID:         "1",
			BodyFn: func() *bytes.Buffer {
				body := `hello world`
				buffer := new(bytes.Buffer)

				_, err := buffer.Write([]byte(body))
				suite.NoError(err)
				return buffer
			},
		},
		{
			Name:           "T22-ZipSaved",
			ExpectedStatus: http.StatusOK,
			TurnID:         "1",
			BodyFn: func() *bytes.Buffer {
				body := `hello world`
				buffer := new(bytes.Buffer)

				zfile := zip.NewWriter(buffer)

				defer zfile.Close()
				w, err := zfile.Create("file")

				suite.NoError(err)
				w.Write([]byte(body))
				return buffer
			},
		},
		{
			Name:           "T23-TurnNotFound",
			ExpectedStatus: http.StatusNotFound,
			TurnID:         "12",
			BodyFn: func() *bytes.Buffer {
				body := `hello world`
				buffer := new(bytes.Buffer)

				zfile := zip.NewWriter(buffer)

				defer zfile.Close()
				w, err := zfile.Create("file")

				suite.NoError(err)
				w.Write([]byte(body))
				return buffer
			},
		},
		{
			Name:           "T24-BadTurnID",
			ExpectedStatus: http.StatusBadRequest,
			TurnID:         "a12",
			BodyFn: func() *bytes.Buffer {
				body := `hello world`
				buffer := new(bytes.Buffer)

				zfile := zip.NewWriter(buffer)

				defer zfile.Close()
				w, err := zfile.Create("file")

				suite.NoError(err)
				w.Write([]byte(body))
				return buffer
			},
		},
	}
	for _, tc := range tcs {
		tc := tc
		suite.T().Run(tc.Name, func(t *testing.T) {

			body := tc.BodyFn()
			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s/files", suite.tServer.URL, tc.TurnID), body)
			suite.NoError(err)
			req.Header.Add("Content-Type", "application/zip")

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

func (m *MockedTurnRepository) Create(request *CreateTurnRequest) (*TurnModel, error) {
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

type MockedMetadataRepository struct {
	mock.Mock
}

func (m *MockedMetadataRepository) Upsert(id int64, path string) error {
	args := m.Called(id)
	return args.Error(0)

}

func (m *MockedMetadataRepository) FindByTurn(id int64) (*MetadataModel, error) {
	args := m.Called(id)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.(*MetadataModel), args.Error(1)

}
