package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const sqlDriverName string = "postgres"

type Storage struct {
	ConnStr string
	db      *sql.DB
}

func NewStorage(connStr string) *Storage {
	return &Storage{
		ConnStr: connStr,
	}
}

func (s *Storage) Init() error {
	if s.ConnStr == "" {
		return fmt.Errorf("connection string not set")
	}

	if err := s.Connect(); err != nil {
		return err
	}

	if err := s.initTables(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Connect() error {
	db, err := sql.Open(sqlDriverName, s.ConnStr)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s *Storage) initTables() error {
	query := `
		create table if not exists users (
			id varchar(50) primary key,
			name varchar(100),
			hashed_password varchar(100),
			created_at timestamp
		)
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *Storage) InsertNewUser(cr UserCreationReq) (*User, error) {
	user, err := NewUser(cr.Name, cr.Password)
	if err != nil {
		return nil, err
	}

	query := `
		insert into users
		(id, name, hashed_password, created_at)
		values
		($1, $2, $3, $4)
	`
	if _, err := s.db.Query(
		query,
		user.ID,
		user.Name,
		user.HashedPassword,
		user.CreatedAt,
	); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Storage) GetAllUsers() ([]*User, error) {
	rows, err := s.db.Query("select * from users")
	if err != nil {
		return nil, err
	}

	users := []*User{}
	for rows.Next() {
		user, err := scanIntoUser(rows)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (s *Storage) GetUserByID(id string) (*User, error) {
	rows, err := s.db.Query("select * from users where id = $1", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoUser(rows)
	}
	return nil, fmt.Errorf("user %s not found", id)
}

func (s *Storage) UpdateUserByID(id string, ur UserUpdateReq) error {
	query := "update users set "
	args := []interface{}{}

	if ur.Name != "" {
		query += "name = $1"
		args = append(args, ur.Name)
	}

	if ur.Password != "" {
		hashedPassword, err := hashPassword(ur.Password)
		if err != nil {
			return err
		}
		if len(args) > 0 {
			query += ", "
		}
		query += "hashed_password = $2"
		args = append(args, hashedPassword)
	}

	if len(args) == 0 {
		return fmt.Errorf("no update fields specified")
	}

	query += " where id = $" + fmt.Sprintf("%d", len(args)+1)
	args = append(args, id)

	_, err := s.db.Exec(query, args...)
	return err
}

func (s *Storage) DeleteUserByID(id string) error {
	_, err := s.db.Query("delete from users where id = $1", id)
	return err
}

func scanIntoUser(rows *sql.Rows) (*User, error) {
	account := new(User)
	err := rows.Scan(
		&account.ID,
		&account.Name,
		&account.HashedPassword,
		&account.CreatedAt,
	)

	return account, err
}

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
