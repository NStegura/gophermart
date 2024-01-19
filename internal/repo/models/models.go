package models

import "time"

type User struct {
	ID        int64
	Login     string
	Password  string
	Salt      string
	CreatedAt time.Time
}
