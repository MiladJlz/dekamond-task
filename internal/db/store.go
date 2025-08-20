package db

import (
	"database/sql"
	"fmt"
	"github.com/MiladJlz/dekamond-task/internal/types"
	_ "github.com/lib/pq"
)

type Store struct {
	DB *sql.DB
}

func NewPostgresDB(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &Store{DB: db}, nil
}

func (s *Store) CreateUser(phone string) error {
	_, err := s.DB.Exec(`INSERT INTO users (phone, created_at) VALUES ($1, NOW())`, phone)
	return err
}

func (s *Store) UserExists(phone string) (bool, error) {
	var exists bool
	err := s.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE phone=$1)`, phone).Scan(&exists)
	return exists, err
}

func (s *Store) GetUserByID(id uint64) (*types.User, error) {
	var user types.User
	err := s.DB.QueryRow(`SELECT id, phone, created_at FROM users WHERE id = $1`, id).Scan(&user.ID, &user.Phone, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByPhone(phone string) (*types.User, error) {
	var user types.User
	err := s.DB.QueryRow(`SELECT * FROM users WHERE phone = $1`, phone).Scan(&user.ID, &user.Phone, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUsers(limit, offset int, search string) ([]types.User, error) {
	query := `SELECT * FROM users`
	args := []any{}
	
	if search != "" {
		query += ` WHERE phone ILIKE $1`
		args = append(args, "%"+search+"%")
	}
	
	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)
		
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []types.User
	for rows.Next() {
		var user types.User
		err := rows.Scan(&user.ID, &user.Phone, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	
	return users, nil
}

func (s *Store) GetUsersCount(search string) (int, error) {
	query := `SELECT COUNT(*) FROM users`
	args := []any{}
	
	if search != "" {
		query += ` WHERE phone ILIKE $1`
		args = append(args, "%"+search+"%")
	}
	
	var count int
	err := s.DB.QueryRow(query, args...).Scan(&count)
	return count, err
}
