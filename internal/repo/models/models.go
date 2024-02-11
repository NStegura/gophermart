package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID        int64
	Login     string
	Password  string
	Balance   float64
	Withdrawn float64
	CreatedAt time.Time
}

type Order struct {
	ID        int64
	Status    string
	UserID    int64
	Accrual   sql.NullFloat64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Withdraw struct {
	ID        int64
	OrderID   int64
	UserID    int64
	Sum       float64
	CreatedAt time.Time
}
