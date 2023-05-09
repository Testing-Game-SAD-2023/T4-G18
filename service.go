package main

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"time"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrBadRequest = errors.New("bad request")
	ErrNotAZip    = errors.New("file is not a valid zip")
)

type GameRepository interface {
	Create(request *CreateGameRequest) (*GameModel, error)
	FindById(id int64) (*GameModel, error)
	Delete(id int64) error
	Update(id int64, ug *UpdateGameRequest) (*GameModel, error)
	FindByInterval(i *IntervalParams, p *PaginationParams) ([]GameModel, int64, error)
	FindByRound(id int64) (*GameModel, error)
}

type GameService struct {
	storage GameRepository
}

func NewGameService(storage GameRepository) *GameService {
	return &GameService{
		storage: storage,
	}
}

func (gc *GameService) Create(request *CreateGameRequest) (*GameModel, error) {
	return gc.storage.Create(request)
}

func (gc *GameService) FindByID(id int64) (*GameModel, error) {
	return gc.storage.FindById(id)
}

func (gc *GameService) Delete(id int64) error {
	return gc.storage.Delete(id)
}

func (gc *GameService) Update(id int64, ug *UpdateGameRequest) (*GameModel, error) {
	return gc.storage.Update(id, ug)
}

func (gc *GameService) FindByInterval(i *IntervalParams, p *PaginationParams) ([]GameModel, int64, error) {
	return gc.storage.FindByInterval(i, p)
}

type RoundRepository interface {
	Create(request *CreateRoundRequest) (*RoundModel, error)
	FindById(id int64) (*RoundModel, error)
	Delete(id int64) error
	Update(id int64, request *UpdateRoundRequest) (*RoundModel, error)
	FindByGame(id int64) ([]RoundModel, error)
}

type RoundService struct {
	storage RoundRepository
}

func NewRoundService(storage RoundRepository) *RoundService {
	return &RoundService{
		storage: storage,
	}
}

func (rs *RoundService) Create(request *CreateRoundRequest) (*RoundModel, error) {
	return rs.storage.Create(request)
}

func (rs *RoundService) Update(id int64, request *UpdateRoundRequest) (*RoundModel, error) {
	return rs.storage.Update(id, request)
}

func (rs *RoundService) FindByID(id int64) (*RoundModel, error) {
	return rs.storage.FindById(id)
}

func (rs *RoundService) FindByGame(id int64) ([]RoundModel, error) {
	return rs.storage.FindByGame(id)
}

func (rs *RoundService) Delete(id int64) error {
	return rs.storage.Delete(id)
}

type TurnRepository interface {
	Create(request *CreateTurnRequest) (*TurnModel, error)
	FindById(id int64) (*TurnModel, error)
	Delete(id int64) error
	Update(id int64, request *UpdateTurnRequest) (*TurnModel, error)
	FindByRound(id int64) ([]TurnModel, error)
}

type MetadataRepository interface {
	Upsert(id int64, path string) error
	FindByTurn(id int64) (*MetadataModel, error)
}

type TurnService struct {
	turnRepository     TurnRepository
	metadataRepository MetadataRepository
	gameRepository     GameRepository
	dataDir            string
}

func NewTurnService(tr TurnRepository, mr MetadataRepository, gr GameRepository, dr string) *TurnService {
	return &TurnService{
		turnRepository:     tr,
		metadataRepository: mr,
		gameRepository:     gr,
		dataDir:            dr,
	}
}

func (ts *TurnService) Create(request *CreateTurnRequest) (*TurnModel, error) {
	return ts.turnRepository.Create(request)
}

func (ts *TurnService) FindByID(id int64) (*TurnModel, error) {
	return ts.turnRepository.FindById(id)
}

func (ts *TurnService) Delete(id int64) error {
	return ts.turnRepository.Delete(id)
}

func (tc *TurnService) FindByRound(id int64) ([]TurnModel, error) {
	return tc.turnRepository.FindByRound(id)
}
func (ts *TurnService) Update(id int64, request *UpdateTurnRequest) (*TurnModel, error) {
	return ts.turnRepository.Update(id, request)
}

func (ts *TurnService) Store(id int64, r io.Reader) error {
	turn, err := ts.turnRepository.FindById(id)
	if err != nil {
		return err
	}
	game, err := ts.gameRepository.FindByRound(turn.RoundID)
	if err != nil {
		return err
	}
	dst, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}
	defer os.Remove(dst.Name())
	if _, err := io.Copy(dst, r); err != nil {
		return err
	}

	if zfile, err := zip.OpenReader(dst.Name()); err != nil {
		return ErrNotAZip
	} else {
		zfile.Close()
	}

	year := time.Now().Year()
	fname := path.Join(ts.dataDir,
		strconv.FormatInt(int64(year), 10),
		strconv.FormatInt(game.ID, 10),
		strconv.FormatInt(id, 10)+".zip",
	)

	dir := path.Dir(fname)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	if err := os.Rename(dst.Name(), fname); err != nil {
		return err
	}

	return ts.metadataRepository.Upsert(id, fname)
}

func (ts *TurnService) GetTurnFile(id int64) (*os.File, error) {
	m, err := ts.metadataRepository.FindByTurn(id)
	if err != nil {
		return nil, err
	}
	return os.Open(m.Path)

}
