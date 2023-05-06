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
	FindById(id uint64) (*GameModel, error)
	Delete(id uint64) error
	Update(id uint64, ug *UpdateGameRequest) (*GameModel, error)
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

func (gc *GameService) FindByID(id uint64) (*GameModel, error) {
	return gc.storage.FindById(id)
}

func (gc *GameService) Delete(id uint64) error {
	return gc.storage.Delete(id)
}

func (gc *GameService) Update(id uint64, ug *UpdateGameRequest) (*GameModel, error) {
	return gc.storage.Update(id, ug)
}

type RoundRepository interface {
	Create(request *CreateRoundRequest) (*RoundModel, error)
	FindById(id uint64) (*RoundModel, error)
	Delete(id uint64) error
	Update(id uint64, request *UpdateRoundRequest) (*RoundModel, error)
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

func (rs *RoundService) Update(id uint64, request *UpdateRoundRequest) (*RoundModel, error) {
	return rs.storage.Update(id, request)
}

func (rs *RoundService) FindByID(id uint64) (*RoundModel, error) {
	return rs.storage.FindById(id)
}

func (rs *RoundService) Delete(id uint64) error {
	return rs.storage.Delete(id)
}

type TurnRepository interface {
	Create(request *CreateTurnRequest) (*TurnModel, error)
	FindById(id uint64) (*TurnModel, error)
	Delete(id uint64) error
	Update(id uint64, request *UpdateTurnRequest) (*TurnModel, error)
	FindGameByTurn(id uint64) (*GameModel, error)
	UpdateMetadata(id uint64, path string) error
	FindMetadataByTurn(id uint64) (*MetadataModel, error)
}

type TurnService struct {
	turnRepository TurnRepository
	dataDir        string
}

func NewTurnService(tr TurnRepository, dr string) *TurnService {
	return &TurnService{
		turnRepository: tr,
		dataDir:        dr,
	}
}

func (ts *TurnService) Create(request *CreateTurnRequest) (*TurnModel, error) {
	return ts.turnRepository.Create(request)
}

func (ts *TurnService) FindByID(id uint64) (*TurnModel, error) {
	return ts.turnRepository.FindById(id)
}

func (ts *TurnService) Delete(id uint64) error {
	return ts.turnRepository.Delete(id)
}

func (ts *TurnService) Update(id uint64, request *UpdateTurnRequest) (*TurnModel, error) {
	return ts.turnRepository.Update(id, request)
}


func (ts *TurnService) Store(turnId uint64, r io.Reader) error {
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

	game, err := ts.turnRepository.FindGameByTurn(turnId)
	if err != nil {
		return err
	}

	year := time.Now().Year()
	fname := path.Join(ts.dataDir,
		strconv.FormatInt(int64(year), 10),
		strconv.FormatUint(game.ID, 10),
		strconv.FormatUint(turnId, 10)+".zip",
	)

	dir := path.Dir(fname)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	if err := os.Rename(dst.Name(), fname); err != nil {
		return err
	}

	return ts.turnRepository.UpdateMetadata(turnId, fname)
}

func (ts *TurnService) GetTurnFile(turnId uint64) (*os.File, error) {
	m, err := ts.turnRepository.FindMetadataByTurn(turnId)
	if err != nil {
		return nil, err
	}
	return os.Open(m.Path)

}
