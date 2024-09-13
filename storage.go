package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
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

// TODO: add createdAt and updatedAt fields to all models in db
var initTablesQueries = []string{
	`create table if not exists users (
		id varchar(50) primary key,
		name varchar(100) unique,
		hashed_password varchar(100),
		created_at timestamp
	)`,
	`create table if not exists email_lists (
		id varchar(50) primary key,
		user_id varchar(50),
		name varchar(100),
		foreign key (user_id) references users(id)
	)`,
	// TODO: add source field to subscriber model
	`create table if not exists subscribers (
		id varchar(50) primary key,
		email_list_id varchar(50),
		user_id varchar(50),
		name varchar(100),
		email_addr varchar(150) unique,
		foreign key (email_list_id) references email_lists(id),
		foreign key (user_id) references users(id)
	)`,
	`create table if not exists outputs (
		id varchar(50) primary key,
		user_id varchar(50),
		output_name varchar(30),
		api_key varchar(100),
		list_id varchar(100),
		foreign key (user_id) references users(id)
	)`,
}

func (s *Storage) initTables() error {
	for _, query := range initTablesQueries {
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) InsertNewUser(cr UserCreationReq) (*User, error) {
	ok, err := validPassword(cr.Password)
	if !ok || err != nil {
		return nil, err
	}

	user, err := NewUser(cr.Name, cr.Password)
	if err != nil {
		return nil, err
	}

	// Preventing root user creds from being saved in db
	if IsRootUser(user) {
		return nil, fmt.Errorf("error creating user")
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

func (s *Storage) GetUserByUsernameAndPassword(username string, password string) (*User, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query("select * from users where name = $1, hashed_password = $2", username, hashedPassword)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoUser(rows)
	}
	return nil, fmt.Errorf("user not found")
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
	user := new(User)
	err := rows.Scan(
		&user.ID,
		&user.Name,
		&user.HashedPassword,
		&user.CreatedAt,
	)
	return user, err
}

func (s *Storage) InsertNewEmailList(cr EmailListCreationReq) (*EmailList, error) {
	emailList := NewEmailList(cr.UserID, cr.Name)

	query := `
		insert into email_lists
		(id, user_id, name)
		values
		($1, $2, $3)
	`
	if _, err := s.db.Query(
		query,
		emailList.ID,
		emailList.UserID,
		emailList.Name,
	); err != nil {
		return nil, err
	}

	return emailList, nil
}

func (s *Storage) GetAllEmailLists() ([]*EmailList, error) {
	rows, err := s.db.Query("select * from email_lists")
	if err != nil {
		return nil, err
	}

	emailLists := []*EmailList{}
	for rows.Next() {
		emailList, err := scanIntoEmailList(rows)
		if err != nil {
			return nil, err
		}

		emailLists = append(emailLists, emailList)
	}

	return emailLists, nil
}

func (s *Storage) GetAllEmailListsByUserID(userID string) ([]*EmailList, error) {
	rows, err := s.db.Query("select * from email_lists where user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	emailLists := []*EmailList{}
	for rows.Next() {
		emailList, err := scanIntoEmailList(rows)
		if err != nil {
			return nil, err
		}

		emailLists = append(emailLists, emailList)
	}

	return emailLists, nil
}

func (s *Storage) GetEmailListByID(id string) (*EmailList, error) {
	rows, err := s.db.Query("select * from email_lists where id = $1", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoEmailList(rows)
	}
	return nil, fmt.Errorf("email list %s not found", id)
}

func (s *Storage) UpdateEmailListByID(id string, ur EmailListUpdateReq) error {
	query := "update email_lists set "
	args := []interface{}{}

	if ur.Name != "" {
		query += "name = $1"
		args = append(args, ur.Name)
	}

	if len(args) == 0 {
		return fmt.Errorf("no update fields specified")
	}

	query += " where id = $" + fmt.Sprintf("%d", len(args)+1)
	args = append(args, id)

	_, err := s.db.Exec(query, args...)
	return err
}

func (s *Storage) DeleteEmailListByID(id string) error {
	_, err := s.db.Query("delete from email_lists where id = $1", id)
	return err
}

func scanIntoEmailList(rows *sql.Rows) (*EmailList, error) {
	emailList := new(EmailList)
	err := rows.Scan(
		&emailList.ID,
		&emailList.UserID,
		&emailList.Name,
	)
	return emailList, err
}

func (s *Storage) InsertNewSubscriber(cr SubscriberCreationReq) (*Subscriber, error) {
	subscriber := NewSubscriber(cr.EmailListID, cr.UserID, cr.Name, cr.EmailAddr)

	query := `
		insert into subscribers
		(id, email_list_id, user_id, name, email_addr)
		values
		($1, $2, $3, $4, $5)
	`
	if _, err := s.db.Query(
		query,
		subscriber.ID,
		subscriber.EmailListID,
		subscriber.UserID,
		subscriber.Name,
		subscriber.EmailAddr,
	); err != nil {
		return nil, err
	}

	return subscriber, nil
}

func (s *Storage) GetAllSubscribers() ([]*Subscriber, error) {
	rows, err := s.db.Query("select * from subscribers")
	if err != nil {
		return nil, err
	}

	subscribers := []*Subscriber{}
	for rows.Next() {
		subscriber, err := scanIntoSubscriber(rows)
		if err != nil {
			return nil, err
		}

		subscribers = append(subscribers, subscriber)
	}

	return subscribers, nil
}

func (s *Storage) GetAllSubscribersByUserID(userID string) ([]*Subscriber, error) {
	rows, err := s.db.Query("select * from subscribers where user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	subscribers := []*Subscriber{}
	for rows.Next() {
		subscriber, err := scanIntoSubscriber(rows)
		if err != nil {
			return nil, err
		}

		subscribers = append(subscribers, subscriber)
	}

	return subscribers, nil
}

func scanIntoSubscriber(rows *sql.Rows) (*Subscriber, error) {
	subscriber := new(Subscriber)
	err := rows.Scan(
		&subscriber.ID,
		&subscriber.EmailListID,
		&subscriber.UserID,
		&subscriber.Name,
		&subscriber.EmailAddr,
	)
	return subscriber, err
}

func (s *Storage) InsertNewOutput(cr OutputCreationReq) (Output, error) {
	id := NewUUID()
	output := makeOutput(id, cr.UserID, cr.OutputName, cr.ApiKey, cr.ListID)

	query := `
		insert into outputs
		(id, user_id, output_name, api_key, list_id)
		values
		($1, $2, $3, $4, $5)
	`
	if _, err := s.db.Query(
		query,
		id,
		cr.UserID,
		output.OutputName(),
		cr.ApiKey,
		cr.ListID,
	); err != nil {
		return nil, err
	}

	return output, nil
}

func (s *Storage) GetAllOutputs() ([]Output, error) {
	outputs := []Output{}
	rows, err := s.db.Query("select * from outputs")
	if err != nil {
		return outputs, err
	}

	for rows.Next() {
		output, err := scanIntoOutput(rows)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}

	return outputs, nil
}

func (s *Storage) GetAllOutputsByUserID(userID string) ([]Output, error) {
	outputs := []Output{}
	rows, err := s.db.Query("select * from outputs where user_id = $1", userID)
	if err != nil {
		return outputs, err
	}

	for rows.Next() {
		output, err := scanIntoOutput(rows)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}

	return outputs, nil
}

func (s *Storage) GetOutputByID(id string) (Output, error) {
	rows, err := s.db.Query("select * from outputs where id = $1", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoOutput(rows)
	}
	return nil, fmt.Errorf("output %s not found", id)
}

func (s *Storage) GetOutputByIDAndUserID(id string, userID string) (Output, error) {
	rows, err := s.db.Query("select * from outputs where id = $1 and user_id = $2", id, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoOutput(rows)
	}
	return nil, fmt.Errorf("output %s not found", id)
}

func scanIntoOutput(rows *sql.Rows) (Output, error) {
	var (
		id         string
		userID     string
		outputName OutputName
		apiKey     string
		listID     string
	)

	err := rows.Scan(
		&id,
		&userID,
		&outputName,
		&apiKey,
		&listID,
	)
	if err != nil {
		return nil, err
	}

	return makeOutput(id, userID, outputName, apiKey, listID), nil
}
