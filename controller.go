package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type CustomInt64 int64

type CustomTime time.Time

type GameService interface {
	Create(request *CreateGameRequest) (*GameModel, error)
	FindById(id int64) (*GameModel, error)
	Delete(id int64) error
	Update(id int64, ug *UpdateGameRequest) (*GameModel, error)
	FindByInterval(i *IntervalParams, p *PaginationParams) ([]GameModel, int64, error)
}
type GameController struct {
	service GameService
}

func NewGameController(gs GameService) *GameController {
	return &GameController{service: gs}
}

func (gc *GameController) create(w http.ResponseWriter, r *http.Request) error {

	request, err := FromJsonBody[CreateGameRequest](r.Body)

	if err != nil {
		return err
	}

	g, err := gc.service.Create(&request)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusCreated, mapToGameDTO(g))

}

func (gc *GameController) findByID(w http.ResponseWriter, r *http.Request) error {

	id, err := FromUrlParams[CustomInt64](r, "id")
	if err != nil {
		return err
	}

	g, err := gc.service.FindById(id.AsInt64())

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, mapToGameDTO(g))

}

func (gc *GameController) delete(w http.ResponseWriter, r *http.Request) error {

	id, err := FromUrlParams[CustomInt64](r, "id")
	if err != nil {
		return err
	}

	if err := gc.service.Delete(id.AsInt64()); err != nil {
		return makeApiError(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (gc *GameController) update(w http.ResponseWriter, r *http.Request) error {

	id, err := FromUrlParams[CustomInt64](r, "id")
	if err != nil {
		return err
	}

	request, err := FromJsonBody[UpdateGameRequest](r.Body)
	if err != nil {
		return err
	}

	g, err := gc.service.Update(id.AsInt64(), &request)
	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, mapToGameDTO(g))
}

type PaginationParams struct {
	page     int64
	pageSize int64
}

type IntervalParams struct {
	startDate time.Time
	endDate   time.Time
}

func (gc *GameController) list(w http.ResponseWriter, r *http.Request) error {
	page, err := FromUrlQuery[CustomInt64](r, "page", 1)

	if err != nil {
		return err
	}

	pageSize, err := FromUrlQuery[CustomInt64](r, "pageSize", 10)

	if err != nil {
		return err
	}

	startDate, err := FromUrlQuery(r, "startDate", CustomTime(time.Now().Add(-24*time.Hour)))

	if err != nil {
		return err
	}

	endDate, err := FromUrlQuery(r, "endDate", CustomTime(time.Now()))

	if err != nil {
		return err
	}

	ip := IntervalParams{
		startDate: startDate.AsTime(),
		endDate:   endDate.AsTime(),
	}

	pp := PaginationParams{
		page:     page.AsInt64(),
		pageSize: pageSize.AsInt64(),
	}

	games, count, err := gc.service.FindByInterval(&ip, &pp)
	if err != nil {
		return makeApiError(err)
	}
	res := make([]*GameDto, len(games))
	for i, game := range games {
		res[i] = mapToGameDTO(&game)
	}

	return writeJson(w, http.StatusOK, makePaginatedResponse(res, count, &pp))
}

type RoundService interface {
	Create(request *CreateRoundRequest) (*RoundModel, error)
	FindById(id int64) (*RoundModel, error)
	Delete(id int64) error
	Update(id int64, request *UpdateRoundRequest) (*RoundModel, error)
	FindByGame(id int64) ([]RoundModel, error)
}

type RoundController struct {
	service RoundService
}

func NewRoundController(rs RoundService) *RoundController {
	return &RoundController{
		service: rs,
	}
}

func (rc *RoundController) create(w http.ResponseWriter, r *http.Request) error {

	request, err := FromJsonBody[CreateRoundRequest](r.Body)
	if err != nil {
		return err
	}

	g, err := rc.service.Create(&request)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusCreated, mapToRoundDTO(g))

}

func (rc *RoundController) update(w http.ResponseWriter, r *http.Request) error {

	id, err := FromUrlParams[CustomInt64](r, "id")
	if err != nil {
		return err
	}

	request, err := FromJsonBody[UpdateRoundRequest](r.Body)
	if err != nil {
		return err
	}

	g, err := rc.service.Update(id.AsInt64(), &request)
	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, mapToRoundDTO(g))
}

func (rc *RoundController) findByID(w http.ResponseWriter, r *http.Request) error {

	id, err := FromUrlParams[CustomInt64](r, "id")
	if err != nil {
		return err
	}

	round, err := rc.service.FindById(id.AsInt64())

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, mapToRoundDTO(round))

}

func (rh *RoundController) delete(w http.ResponseWriter, r *http.Request) error {

	id, err := FromUrlParams[CustomInt64](r, "id")
	if err != nil {
		return err
	}

	if err := rh.service.Delete(id.AsInt64()); err != nil {
		return makeApiError(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (rc *RoundController) list(w http.ResponseWriter, r *http.Request) error {
	id, err := FromUrlQuery(r, "gameId", CustomInt64(10))

	if err != nil {
		return err
	}

	rounds, err := rc.service.FindByGame(id.AsInt64())
	if err != nil {
		return makeApiError(err)
	}

	resp := make([]*RoundDto, len(rounds))
	for i, round := range rounds {
		resp[i] = mapToRoundDTO(&round)
	}

	return writeJson(w, http.StatusOK, resp)
}

type TurnService interface {
	CreateBulk(request *CreateTurnsRequest) ([]TurnModel, error)
	FindById(id int64) (*TurnModel, error)
	Delete(id int64) error
	Update(id int64, request *UpdateTurnRequest) (*TurnModel, error)
	FindByRound(id int64) ([]TurnModel, error)
	SaveFile(id int64, r io.Reader) error
	GetFile(id int64) (string, *os.File, error)
}

type TurnController struct {
	service TurnService
}

func NewTurnController(service TurnService) *TurnController {
	return &TurnController{
		service: service,
	}
}

func (tc *TurnController) create(w http.ResponseWriter, r *http.Request) error {

	request, err := FromJsonBody[CreateTurnsRequest](r.Body)
	if err != nil {
		return err
	}
	turns, err := tc.service.CreateBulk(&request)

	if err != nil {
		return makeApiError(err)
	}

	resp := make([]*TurnDto, len(turns))
	for i, turn := range turns {
		resp[i] = mapToTurnDTO(&turn)
	}

	return writeJson(w, http.StatusCreated, resp)
}

func (tc *TurnController) update(w http.ResponseWriter, r *http.Request) error {

	id, err := FromUrlParams[CustomInt64](r, "id")
	if err != nil {
		return err
	}

	request, err := FromJsonBody[UpdateTurnRequest](r.Body)
	if err != nil {
		return err
	}

	g, err := tc.service.Update(id.AsInt64(), &request)
	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, mapToTurnDTO(g))
}

func (tc *TurnController) findByID(w http.ResponseWriter, r *http.Request) error {

	id, err := FromUrlParams[CustomInt64](r, "id")
	if err != nil {
		return err
	}

	turn, err := tc.service.FindById(id.AsInt64())

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, mapToTurnDTO(turn))

}

func (tc *TurnController) delete(w http.ResponseWriter, r *http.Request) error {

	id, err := FromUrlParams[CustomInt64](r, "id")
	if err != nil {
		return err
	}

	if err := tc.service.Delete(id.AsInt64()); err != nil {
		return makeApiError(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (tc *TurnController) upload(w http.ResponseWriter, r *http.Request) error {

	id, err := FromUrlParams[CustomInt64](r, "id")
	if err != nil {
		return err
	}

	if err := tc.service.SaveFile(id.AsInt64(), r.Body); err != nil {
		return makeApiError(err)
	}
	defer r.Body.Close()

	w.WriteHeader(http.StatusOK)
	return nil
}

func (tc *TurnController) download(w http.ResponseWriter, r *http.Request) error {
	id, err := FromUrlParams[CustomInt64](r, "id")
	if err != nil {
		return err
	}

	fname, f, err := tc.service.GetFile(id.AsInt64())
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
	id, err := FromUrlQuery(r, "roundId", CustomInt64(10))

	if err != nil {
		return err
	}
	turns, err := tc.service.FindByRound(id.AsInt64())
	if err != nil {
		return makeApiError(err)
	}

	resp := make([]*TurnDto, len(turns))
	for i, turn := range turns {
		resp[i] = mapToTurnDTO(&turn)
	}

	return writeJson(w, http.StatusOK, resp)
}

func (CustomTime) Convert(s string) (CustomTime, error) {
	t, err := time.Parse("2006-01-02", s)
	return CustomTime(t), err
}

func (CustomTime) Validate() error {
	return nil
}

func (k CustomTime) AsTime() time.Time {
	return time.Time(k)
}

func (CustomInt64) Convert(s string) (CustomInt64, error) {
	a, err := strconv.ParseInt(s, 10, 64)
	return CustomInt64(a), err
}

func (CustomInt64) Validate() error {
	return nil
}

func (k CustomInt64) AsInt64() int64 {
	return int64(k)
}
