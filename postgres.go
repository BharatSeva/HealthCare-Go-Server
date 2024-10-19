package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)
type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(url string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.CreateTable()
}

func (s *PostgresStore) CreateTable() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS user_table (
		id serial primary key,
		name varchar(100),
		email varchar(100)
		)`,
	}
	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) SignUp(user *User) error {
	query := `insert into user_table (name, email) values($1, $2)`
	_, err := s.db.Query(query, user.Name, user.Email)
	if err != nil {
		return err
	}
	return nil
}

// Login with "email" and "password" to generate JWT token
func (s *PostgresStore) LoginUser(acc *Login) (*User, error) {
	var user User
	err := s.db.QueryRow("select * from user_table where email = $1", acc.Email).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return nil, fmt.Errorf("no account found with given email: %s", acc.Email)
	}
	return &user, nil
}