package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

func gameModelToDto(g *GameModel) *GameDto {
	return &GameDto{
		ID:           g.ID,
		CurrentRound: g.CurrentRound,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
		PlayersCount: g.PlayersCount,
	}
}

func roundModelToDto(g *RoundModel) *RoundDto {
	return &RoundDto{
		ID:          g.ID,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
		IdTestClass: g.IdTestClass,
	}
}

func turnModelToDto(t *TurnModel) *TurnDto {
	return &TurnDto{
		ID:          t.ID,	
		IsWinner:  	 t.IsWinner,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,	
		PlayerID:    t.PlayerID,
	}
}

func writeJson(w http.ResponseWriter, statusCode int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(v)
}

func makeApiError(err error) error {

	switch {
	case errors.Is(err, ErrNotFound):
		return ApiError{code: http.StatusNotFound, Message: "Resource not found"}
	case errors.Is(err, ErrBadRequest):
		return ApiError{code: http.StatusBadRequest, Message: "Bad request"}
	default:
		return ApiError{code: http.StatusInternalServerError, Message: "Internal server error"}
	}
}
