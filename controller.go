package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type GameController struct {
	service *GameService
}

func NewGameController(gc *GameService) *GameController {
	return &GameController{service: gc}
}

func (gh *GameController) create(w http.ResponseWriter, r *http.Request) error {

	var request CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid json body",
		}
	}

	defer r.Body.Close()

	g, err := gh.service.Create(&request)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusCreated, gameModelToDto(g))

}

func (gh *GameController) findByID(w http.ResponseWriter, r *http.Request) error {

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid game id",
		}
	}
	g, err := gh.service.FindByID(id)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, gameModelToDto(g))

}

func (gh *GameController) delete(w http.ResponseWriter, r *http.Request) error {

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid game id",
		}
	}

	if err := gh.service.Delete(id); err != nil {
		return makeApiError(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type RoundController struct {
	service *RoundService
}

func NewRoundController(rs *RoundService) *RoundController {
	return &RoundController{
		service: rs,
	}
}

func (rh *RoundController) create(w http.ResponseWriter, r *http.Request) error {

	var request CreateRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid json body",
		}
	}

	defer r.Body.Close()

	g, err := rh.service.Create(&request)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusCreated, roundModelToDto(g))

}

func (rh *RoundController) findByID(w http.ResponseWriter, r *http.Request) error {

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid game id",
		}
	}
	round, err := rh.service.FindByID(id)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, roundModelToDto(round))

}

func (rh *RoundController) delete(w http.ResponseWriter, r *http.Request) error {

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid round id",
		}
	}

	if err := rh.service.Delete(id); err != nil {
		return makeApiError(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type TurnController struct {
	service *TurnService
}

func NewTurnController(service *TurnService) *TurnController {
	return &TurnController{
		service: service,
	}
}

func (tc *TurnController) upload(w http.ResponseWriter, r *http.Request) error {

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid round id",
		}
	}
	if err := tc.service.Store(id, r.Body); err != nil {
		return makeApiError(err)
	}
	return nil
}

func (tc *TurnController) download(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid round id",
		}
	}

	f, err := tc.service.GetTurnFile(id)
	if err != nil {
		return makeApiError(err)
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/zip")
	b := make([]byte, 4096)
	for {
		n, err := f.Read(b)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		if _, err := w.Write(b[:n]); err != nil {
			return err
		}
	}
	return nil
}