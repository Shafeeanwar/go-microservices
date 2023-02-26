package data

import (
	"database/sql"
	"time"
)

type PostgresTestRepository struct {
	Conn *sql.DB
}

func NewPostgresTestRepository(db *sql.DB) *PostgresTestRepository {
	return &PostgresTestRepository{
		Conn: db,
	}
}

func (u *PostgresTestRepository) GetAll() ([]*User, error) {
	users := []*User{}

	return users, nil
}

func (u *PostgresTestRepository) GetByEmail(email string) (*User, error) {
	user := User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "A",
		LastName:  "Z",
		Password:  "P",
		Active:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return &user, nil
}

func (u *PostgresTestRepository) GetOne(id int) (*User, error) {
	user := User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "A",
		LastName:  "Z",
		Password:  "P",
		Active:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return &user, nil
}

func (u *PostgresTestRepository) Update(user User) error {
	return nil
}

func (u *PostgresTestRepository) DeleteByID(id int) error {
	return nil
}

func (u *PostgresTestRepository) Insert(user User) (int, error) {
	return 1, nil
}

func (u *PostgresTestRepository) ResetPassword(password string, user User) error {
	return nil
}

func (u *PostgresTestRepository) PasswordMatches(plainText string, user User) (bool, error) {
	return true, nil
}
