package main

import (
	"context"
	"net/http"
	"time"

	"github.com/guionardo/typedhandler/typedhandler"
)

type (
	LoginRequest struct {
		UserName string `json:"username"`
		Password string `json:"password"`
	}
	LoginResponse struct {
		UserName   string    `json:"username"`
		Token      string    `json:"token"`
		ValidUntil time.Time `json:"valid_until"`
	}
)

func main() {
	requestParser := typedhandler.CreateParser[*LoginRequest]()
	handler := typedhandler.CreateHandler(requestParser, serviceFunc)
	http.HandleFunc("POST /login", handler)

	if err := http.ListenAndServe(":8000", http.DefaultServeMux); err != nil { //nolint: gosec
		panic(err)
	}
}

func serviceFunc(ctx context.Context, request *LoginRequest) (LoginResponse, int, error) {
	return LoginResponse{
		UserName: request.UserName,
		Token:    "abcd", ValidUntil: time.Now().AddDate(0, 0, 1)}, http.StatusOK, nil
}
