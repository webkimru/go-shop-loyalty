package models

type User struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
}
