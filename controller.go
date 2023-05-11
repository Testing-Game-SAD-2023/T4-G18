package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type GameController struct {
	service *GameService
}

func NewGameController(gc *GameService) *GameController {
	return &GameController{service: gc}
}

func (gc *GameController) create(w http.ResponseWriter, r *http.Request) error {

	var request CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid json body",
		}
	}

	defer r.Body.Close()

	g, err := gc.service.Create(&request)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusCreated, mapToGameDTO(g))

}

func (gc *GameController) findByID(w http.ResponseWriter, r *http.Request) error {

	id := r.Context().Value(idParamKey).(int64)

	g, err := gc.service.FindByID(id)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, mapToGameDTO(g))

}

func (gc *GameController) delete(w http.ResponseWriter, r *http.Request) error {

	id := r.Context().Value(idParamKey).(int64)

	if err := gc.service.Delete(id); err != nil {
		return makeApiError(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (gc *GameController) update(w http.ResponseWriter, r *http.Request) error {

	id := r.Context().Value(idParamKey).(int64)

	var rq UpdateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&rq); err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid json body",
		}
	}
	defer r.Body.Close()

	g, err := gc.service.Update(id, &rq)
	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, mapToGameDTO(g))
}

func (gc *GameController) list(w http.ResponseWriter, r *http.Request) error {
	paginationParams := r.Context().Value(paginationParamsKey).(PaginationParams)
	intervalParams := r.Context().Value(intervalParamsKey).(IntervalParams)

	games, count, err := gc.service.FindByInterval(&intervalParams, &paginationParams)
	if err != nil {
		return makeApiError(err)
	}
	res := make([]*GameDto, len(games))
	for i, game := range games {
		res[i] = mapToGameDTO(&game)
	}

	return writeJson(w, http.StatusOK, makePaginatedResponse(res, count, &paginationParams))
}

type RoundController struct {
	service *RoundService
}

func NewRoundController(rs *RoundService) *RoundController {
	return &RoundController{
		service: rs,
	}
}

func (rc *RoundController) create(w http.ResponseWriter, r *http.Request) error {

	var request CreateRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid json body",
		}
	}

	defer r.Body.Close()

	g, err := rc.service.Create(&request)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusCreated, mapToRoundDTO(g))

}

func (rc *RoundController) update(w http.ResponseWriter, r *http.Request) error {

	id := r.Context().Value(idParamKey).(int64)

	var rq UpdateRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&rq); err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid json body",
		}
	}
	defer r.Body.Close()

	g, err := rc.service.Update(id, &rq)
	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, mapToRoundDTO(g))
}

func (rc *RoundController) findByID(w http.ResponseWriter, r *http.Request) error {

	id := r.Context().Value(idParamKey).(int64)

	round, err := rc.service.FindByID(id)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, mapToRoundDTO(round))

}

func (rh *RoundController) delete(w http.ResponseWriter, r *http.Request) error {

	id := r.Context().Value(idParamKey).(int64)

	if err := rh.service.Delete(id); err != nil {
		return makeApiError(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (rc *RoundController) list(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(r.URL.Query().Get("gameId"), 10, 64)

	if err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid game id",
		}
	}

	rounds, err := rc.service.FindByGame(id)
	if err != nil {
		return makeApiError(err)
	}

	resp := make([]*RoundDto, len(rounds))
	for i, round := range rounds {
		resp[i] = mapToRoundDTO(&round)
	}

	return writeJson(w, http.StatusOK, resp)
}

type TurnController struct {
	service *TurnService
}

func NewTurnController(service *TurnService) *TurnController {
	return &TurnController{
		service: service,
	}
}

func (tc *TurnController) create(w http.ResponseWriter, r *http.Request) error {

	var request CreateTurnRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid json body",
		}
	}

	defer r.Body.Close()

	g, err := tc.service.Create(&request)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusCreated, mapToTurnDTO(g))
}

func (tc *TurnController) update(w http.ResponseWriter, r *http.Request) error {

	id := r.Context().Value(idParamKey).(int64)

	var rq UpdateTurnRequest
	if err := json.NewDecoder(r.Body).Decode(&rq); err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid json body",
		}
	}
	defer r.Body.Close()

	g, err := tc.service.Update(id, &rq)
	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, mapToTurnDTO(g))
}

func (tc *TurnController) findByID(w http.ResponseWriter, r *http.Request) error {

	id := r.Context().Value(idParamKey).(int64)

	turn, err := tc.service.FindByID(id)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, mapToTurnDTO(turn))

}

func (tc *TurnController) delete(w http.ResponseWriter, r *http.Request) error {

	id := r.Context().Value(idParamKey).(int64)

	if err := tc.service.Delete(id); err != nil {
		return makeApiError(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (tc *TurnController) upload(w http.ResponseWriter, r *http.Request) error {

	id := r.Context().Value(idParamKey).(int64)
	defer r.Body.Close()

	if err := tc.service.Store(id, r.Body); err != nil {
		return makeApiError(err)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (tc *TurnController) download(w http.ResponseWriter, r *http.Request) error {
	id := r.Context().Value(idParamKey).(int64)

	fname, f, err := tc.service.GetTurnFile(id)
	if err != nil {
		return makeApiError(err)
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fname))
	if _, err := io.Copy(w, f); err != nil {
		return err
	}
	return nil
}

func (tc *TurnController) list(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(r.URL.Query().Get("roundId"), 10, 64)

	if err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid round id",
		}
	}

	turns, err := tc.service.FindByRound(id)
	if err != nil {
		return makeApiError(err)
	}

	resp := make([]*TurnDto, len(turns))
	for i, turn := range turns {
		resp[i] = mapToTurnDTO(&turn)
	}

	return writeJson(w, http.StatusOK, resp)
}
