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
	ID        int64
	Status    string
	UserID    int64
	CreatedAt time.Time
}
