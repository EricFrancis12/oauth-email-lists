package main

import (
	"database/sql"
	"fmt"
	"time"

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

func sqlTrigger(triggerName string, tableName string) string {
	return fmt.Sprintf(
		"create or replace trigger %s before update on %s for each row execute procedure update_modified_column();",
		triggerName,
		tableName,
	)
}

var initTablesQueries = []string{
	`create or replace function update_modified_column()
		returns trigger as $$
		begin
		new.updated_at = now();
		return new;
		end;
		$$ language 'plpgsql';
	`,
	`create table if not exists users (
		id varchar(50) primary key,
		name varchar(100) unique,
		hashed_password varchar(100),
		created_at timestamp default current_timestamp,
		updated_at timestamp default current_timestamp
	)`,
	sqlTrigger("update_users_updated_at", "users"),
	`create table if not exists email_lists (
		id varchar(50) primary key,
		user_id varchar(50),
		name varchar(100),
		created_at timestamp default current_timestamp,
		updated_at timestamp default current_timestamp,
		foreign key (user_id) references users(id)
	)`,
	sqlTrigger("update_email_lists_updated_at", "email_lists"),
	`create table if not exists subscribers (
		id varchar(50) primary key,
		email_list_id varchar(50),
		user_id varchar(50),
		source_provider_name varchar(50),
		name varchar(100),
		email_addr varchar(150) unique,
		created_at timestamp default current_timestamp,
		updated_at timestamp default current_timestamp,
		foreign key (email_list_id) references email_lists(id),
		foreign key (user_id) references users(id)
	)`,
	sqlTrigger("update_subscribers_updated_at", "subscribers"),
	`create table if not exists outputs (
		id varchar(50) primary key,
		user_id varchar(50),
		output_name varchar(30),
		list_id varchar(100),
		created_at timestamp default current_timestamp,
		updated_at timestamp default current_timestamp,
		foreign key (user_id) references users(id)
	)`,
	sqlTrigger("update_outputs_updated_at", "outputs"),
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
		&user.UpdatedAt,
	)
	return user, err
}

func (s *Storage) InsertNewEmailList(cr EmailListCreationReq) (*EmailList, error) {
	emailList := NewEmailList(cr.UserID, cr.Name)

	query := `
		insert into email_lists
		(id, user_id, name, created_at, updated_at)
		values
		($1, $2, $3, $4, $5)
	`
	if _, err := s.db.Query(
		query,
		emailList.ID,
		emailList.UserID,
		emailList.Name,
		emailList.CreatedAt,
		emailList.UpdatedAt,
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
		&emailList.CreatedAt,
		&emailList.UpdatedAt,
	)
	return emailList, err
}

func (s *Storage) InsertNewSubscriber(cr SubscriberCreationReq) (*Subscriber, error) {
	subscriber := NewSubscriber(cr.EmailListID, cr.UserID, cr.SourceProviderName, cr.Name, cr.EmailAddr)

	query := `
		insert into subscribers
		(id, email_list_id, user_id, source_provider_name, name, email_addr, created_at, updated_at)
		values
		($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if _, err := s.db.Query(
		query,
		subscriber.ID,
		subscriber.EmailListID,
		subscriber.UserID,
		subscriber.SourceProviderName,
		subscriber.Name,
		subscriber.EmailAddr,
		subscriber.CreatedAt,
		subscriber.UpdatedAt,
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
		&subscriber.SourceProviderName,
		&subscriber.Name,
		&subscriber.EmailAddr,
		&subscriber.CreatedAt,
		&subscriber.UpdatedAt,
	)
	return subscriber, err
}

func (s *Storage) InsertNewOutput(cr OutputCreationReq) (Output, error) {
	id := NewUUID()
	now := time.Now()
	output := makeOutput(id, cr.UserID, cr.OutputName, cr.ListID, now, now)

	query := `
		insert into outputs
		(id, user_id, output_name, list_id, created_at, updated_at)
		values
		($1, $2, $3, $4, $5, $6)
	`
	if _, err := s.db.Query(
		query,
		id,
		cr.UserID,
		output.OutputName(),
		cr.ListID,
		now,
		now,
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

func (s *Storage) UpdateOutputByIDAndUserID(id string, userID string, ur OutputUpdateReq) error {
	query := "update outputs set "
	args := []interface{}{}

	if ur.OutputName != "" {
		query += "output_name = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, ur.OutputName)
	}
	if ur.ListID != "" {
		query += "list_id = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, ur.ListID)
	}

	if len(args) == 0 {
		return fmt.Errorf("no update fields specified")
	}

	query += " where id = $" + fmt.Sprintf("%d", len(args)+1)
	query += " and user_id = $" + fmt.Sprintf("%d", len(args)+2)
	args = append(args, id, userID)

	_, err := s.db.Exec(query, args...)
	return err
}

func scanIntoOutput(rows *sql.Rows) (Output, error) {
	var (
		id         string
		userID     string
		outputName OutputName
		listID     string
		createdAt  time.Time
		updatedAt  time.Time
	)

	err := rows.Scan(
		&id,
		&userID,
		&outputName,
		&listID,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	return makeOutput(id, userID, outputName, listID, createdAt, updatedAt), nil
}
