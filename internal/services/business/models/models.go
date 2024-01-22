package models

import "time"

type User struct {
	ID        int64
	Login     string
	Password  string
	Salt      string
	CreatedAt time.Time
}

type Order struct {
	Number     int64
	Status     string
	Accrual    float64
	UploadedAt time.Time
}
