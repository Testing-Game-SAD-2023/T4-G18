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

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *GameControllerSuite) SetupSuite() {
	gr := new(MockedGameRepository)
	gr.
		On("Create", &CreateGameRequest{PlayersCount: 10}).Return(&GameModel{ID: 1}, nil).
		On("FindById", uint64(1)).Return(&GameModel{ID: 1}, nil).
		On("FindById", mock.MatchedBy(func(id uint64) bool { return id != 1 })).Return(nil, ErrNotFound).
		On("Delete", uint64(1)).Return(nil).
		On("Delete", mock.Anything).Return(ErrNotFound)

	service := NewGameService(gr)
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

func (gr *MockedGameRepository) FindById(id uint64) (*GameModel, error) {
	args := gr.Called(id)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.(*GameModel), args.Error(1)

}

func (gr *MockedGameRepository) Delete(id uint64) error {
	args := gr.Called(id)
	return args.Error(0)
}

type TurnControllerSuite struct {
	suite.Suite
	tServer *httptest.Server
	tmpDir  string
}

func (suite *TurnControllerSuite) SetupSuite() {
	tr := new(MockedTurnRepository)
	f ,err := os.CreateTemp(suite.tmpDir, "")
	f.Write([]byte("hello"))
	defer f.Close()
	suite.NoError(err)
	tr.
		On("FindGameByTurn", uint64(1)).Return(&GameModel{ID: 1}, nil).
		On("FindGameByTurn", mock.MatchedBy(func(id uint64) bool { return id != 1 })).Return(nil, ErrNotFound).
		On("UpdateMetadata", uint64(1)).Return(nil).
		On("UpdateMetadata", mock.MatchedBy(func(id uint64) bool { return id != 1 })).Return(ErrNotFound).
		On("FindMetadataByTurn", uint64(1)).Return(&MetadataModel{Path: f.Name()}, nil).
		On("FindMetadataByTurn", mock.MatchedBy(func(id uint64) bool { return id != 1 })).Return(nil, ErrNotFound)

	suite.tmpDir = os.TempDir()
	service := NewTurnService(tr, suite.tmpDir)
	controller := NewTurnController(service)

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
			Name:           "T31-BadZip",
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
			Name:           "T32-ZipSaved",
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
			Name:           "T33-TurnNotFound",
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
			Name:           "T34-BadTurnID",
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
			suite.Equal(tc.ExpectedStatus, res.StatusCode)

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
			suite.Equal(tc.ExpectedStatus, res.StatusCode)

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

func (m *MockedTurnRepository) FindGameByTurn(id uint64) (*GameModel, error) {
	args := m.Called(id)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.(*GameModel), args.Error(1)

}
func (m *MockedTurnRepository) UpdateMetadata(id uint64, path string) error {
	args := m.Called(id)
	return args.Error(0)

}

func (m *MockedTurnRepository) FindMetadataByTurn(id uint64) (*MetadataModel, error) {
	args := m.Called(id)
	v := args.Get(0)

	if v == nil {
		return nil, args.Error(1)
	}
	return v.(*MetadataModel), args.Error(1)

}
