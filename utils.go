package main

import (
	"encoding/json"
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

func writeJson(w http.ResponseWriter, statusCode int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return err
	}
	return nil
}
