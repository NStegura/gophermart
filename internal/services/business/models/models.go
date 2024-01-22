package models

import "time"

type User struct {
	ID        int64
	Login     string
	Password  string
	Salt      string
	Balance   float64
	Withdrawn float64
	CreatedAt time.Time
}

type Order struct {
	Number     int64
	Status     string
	Accrual    float64
	UploadedAt time.Time
}

type Withdraw struct {
	OrderID   int64
	Sum       float64
	CreatedAt time.Time
}
