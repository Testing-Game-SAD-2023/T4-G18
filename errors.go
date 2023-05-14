package main

import (
	"errors"
	"net/http"

	"gorm.io/gorm"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrBadRequest        = errors.New("bad request")
	ErrNotAZip           = errors.New("file is not a valid zip")
	ErrInvalidRoundOrder = errors.New("invalid round order")
	ErrDuplicateKey      = errors.New("duplicated key")
	ErrInvalidPlayerList = errors.New("invalid player list")
	ErrInvalidParam      = errors.New("invalid param")
)

func handleError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return ErrDuplicateKey
	default:
		return err
	}
}

func makeApiError(err error) error {
	var code int

	switch {
	case errors.Is(err, ErrNotFound):
		code = http.StatusNotFound
	case errors.Is(err, ErrBadRequest),
		errors.Is(err, ErrInvalidPlayerList),
		errors.Is(err, ErrInvalidParam),
		errors.Is(err, ErrInvalidRoundOrder):
		code = http.StatusBadRequest
	case errors.Is(err, ErrNotAZip):
		code = http.StatusUnprocessableEntity
	case errors.Is(err, ErrDuplicateKey):
		code = http.StatusConflict
	default:
		code = http.StatusInternalServerError
	}

	if _, ok := err.(*http.MaxBytesError); ok {
		code = http.StatusRequestEntityTooLarge
	}

	return ApiError{code: code, Message: err.Error(), err: err}
}
