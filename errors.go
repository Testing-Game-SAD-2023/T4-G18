package main

import (
	"errors"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrBadRequest        = errors.New("bad request")
	ErrNotAZip           = errors.New("file is not a valid zip")
	ErrInvalidRoundOrder = errors.New("invalid round order")
	ErrDuplicatedKey     = errors.New("already exists")
	ErrInvalidPlayerList = errors.New("invalid player list")
	ErrInvalidParam      = errors.New("invalid param")
)

func handleError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return ErrDuplicatedKey
	default:
		return err
	}
}

func makeApiError(err error) error {
	var code int
	var message string

	switch {
	case errors.Is(err, ErrNotFound):
		code = http.StatusNotFound
		message = err.Error()
	case errors.Is(err, ErrBadRequest),
		errors.Is(err, ErrInvalidPlayerList),
		errors.Is(err, ErrInvalidParam),
		errors.Is(err, ErrInvalidRoundOrder):
		code = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, ErrNotAZip):
		code = http.StatusUnprocessableEntity
		message = err.Error()
	case errors.Is(err, ErrDuplicatedKey):
		code = http.StatusConflict
		message = err.Error()
	default:
		if err, ok := err.(*http.MaxBytesError); ok {
			code = http.StatusRequestEntityTooLarge
			message = fmt.Sprintf("allowed body size: %s", ByteCountIEC(err.Limit))

		} else {
			code = http.StatusInternalServerError
			message = "internal server error"
		}
	}

	return ApiError{code: code, Message: message, err: err}
}
