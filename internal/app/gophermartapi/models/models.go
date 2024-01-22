package models

import "time"

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     int64     `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawIn struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type WithdrawOut struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
