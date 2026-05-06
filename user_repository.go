package main

import (
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("Invalid Credentials!")

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (u *UserRepo) InsertUser(name, email, plainPassword, avatar string) (int, error) {

	ctx := context.Background()

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	// after commit there is no effect on rollback, so if transaction fails before commit then it's rolled back
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `insert into users (name, email, hashed_password) values (?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	res, err := stmt.Exec(name, email, string(hashedPassword))
	if err != nil {
		return 0, err
	}

	userId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	profileStmt, err := tx.PrepareContext(ctx, `insert into profile (user_id, avatar) values (?, ?)`)
	if err != nil {
		return 0, err
	}
	defer profileStmt.Close()

	_, err = profileStmt.Exec(userId, avatar)
	if err != nil {
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return int(userId), nil
}

func (u *UserRepo) FetchUser(email string) (*User, error) {
	var query = `select u.id, u.name, u.email, u.hashed_password, u.created_at, p.avatar from users u
	INNER JOIN profile p ON u.id = p.user_id where email = ?`

	row := u.db.QueryRow(query, email)
	var user User

	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.HashedPassword, &user.CreatedAt, &user.Profile.Avatar)
	if err != nil {
		return nil, err
	}
	user.Profile.UserId = user.Id
	return &user, nil
}

func (u *UserRepo) Authenticate(email string, plainPassword string) (string, error) {
	user, err := u.FetchUser(email)
	if err != nil {
		return "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(plainPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	return user.Email, nil
}
