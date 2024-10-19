package main

// import (
// 	"golang.org/x/crypto/bcrypt"
// )

type User struct {
	ID       int    `db:"id" json:"id"`
	Name string `json:"name"`
	Email    string `json:"email"`
}


type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}