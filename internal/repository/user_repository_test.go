package repository

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/DrusGalkin/forum-auth-grpc/pkg/logger"
	"go.uber.org/zap"
	"testing"

	"github.com/DrusGalkin/forum-auth-grpc/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger.Logger = zap.NewNop()
	repo := NewUserRepository(db)

	tests := []struct {
		name    string
		mock    func()
		want    []entity.User
		wantErr bool
	}{
		{
			name: "Success",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role"}).
					AddRow(1, "User1", "user1@test.com", "pass1", "user").
					AddRow(2, "User2", "user2@test.com", "pass2", "admin")
				mock.ExpectQuery("SELECT id, name, email, password, role FROM users ORDER BY id").WillReturnRows(rows)
			},
			want: []entity.User{
				{ID: 1, Name: "User1", Email: "user1@test.com", Password: "pass1", Role: "user"},
				{ID: 2, Name: "User2", Email: "user2@test.com", Password: "pass2", Role: "admin"},
			},
			wantErr: false,
		},
		{
			name: "Empty result",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role"})
				mock.ExpectQuery("SELECT id, name, email, password, role FROM users ORDER BY id").WillReturnRows(rows)
			},
			want:    []entity.User(nil),
			wantErr: false,
		},
		{
			name: "Query error",
			mock: func() {
				mock.ExpectQuery("SELECT id, name, email, password, role FROM users ORDER BY id").WillReturnError(errors.New("query error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.GetAll()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger.Logger = zap.NewNop()
	repo := NewUserRepository(db)

	tests := []struct {
		name    string
		id      int
		mock    func()
		want    entity.User
		wantErr error
	}{
		{
			name: "Success",
			id:   1,
			mock: func() {
				row := sqlmock.NewRows([]string{"id", "name", "email", "password", "role"}).
					AddRow(1, "User1", "user1@test.com", "pass1", "user")
				mock.ExpectQuery("SELECT id, name, email, password, role FROM users WHERE id =").WithArgs(1).WillReturnRows(row)
			},
			want:    entity.User{ID: 1, Name: "User1", Email: "user1@test.com", Password: "pass1", Role: "user"},
			wantErr: nil,
		},
		{
			name: "Not found",
			id:   999,
			mock: func() {
				mock.ExpectQuery("SELECT id, name, email, password, role FROM users WHERE id =").WithArgs(999).WillReturnError(sql.ErrNoRows)
			},
			want:    entity.User{},
			wantErr: entity.ErrorUserNotFound,
		},
		{
			name: "Database error",
			id:   1,
			mock: func() {
				mock.ExpectQuery("SELECT id, name, email, password, role FROM users WHERE id =").WithArgs(1).WillReturnError(errors.New("db error"))
			},
			want:    entity.User{},
			wantErr: errors.New("ошибка при получении пользователя по ID: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.GetByID(tt.id)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger.Logger = zap.NewNop()
	repo := NewUserRepository(db)

	user := entity.User{
		Name:     "NewUser",
		Email:    "new@test.com",
		Password: "newpass",
	}

	tests := []struct {
		name    string
		mock    func()
		want    entity.User
		wantErr bool
	}{
		{
			name: "Success",
			mock: func() {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs(user.Name, user.Email, user.Password).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password"}).
						AddRow(1, user.Name, user.Email, user.Password))
			},
			want:    entity.User{ID: 1, Name: user.Name, Email: user.Email, Password: user.Password},
			wantErr: false,
		},
		{
			name: "Duplicate email",
			mock: func() {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs(user.Name, user.Email, user.Password).
					WillReturnError(errors.New("duplicate key value violates unique constraint"))
			},
			want:    entity.User{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.Create(user)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger.Logger = zap.NewNop()
	repo := NewUserRepository(db)

	user := entity.User{
		Name:     "UpdatedUser",
		Email:    "updated@test.com",
		Password: "updatedpass",
	}

	tests := []struct {
		name    string
		id      int
		mock    func()
		want    entity.User
		wantErr error
	}{
		{
			name: "Not found",
			id:   999,
			mock: func() {
				mock.ExpectQuery("UPDATE users").
					WithArgs(user.Name, user.Email, user.Password, 999).
					WillReturnError(sql.ErrNoRows)
			},
			want:    entity.User{},
			wantErr: entity.ErrorUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := repo.Update(tt.id, user)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger.Logger = zap.NewNop()
	repo := NewUserRepository(db)

	tests := []struct {
		name    string
		id      int
		mock    func()
		wantErr error
	}{
		{
			name: "Success",
			id:   1,
			mock: func() {
				mock.ExpectExec("DELETE FROM users WHERE id = ").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: nil,
		},
		{
			name: "Not found",
			id:   999,
			mock: func() {
				mock.ExpectExec("DELETE FROM users WHERE id = ").
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: entity.ErrorUserNotFound,
		},
		{
			name: "Database error",
			id:   1,
			mock: func() {
				mock.ExpectExec("DELETE FROM users WHERE id = ").
					WithArgs(1).
					WillReturnError(errors.New("db error"))
			},
			wantErr: errors.New("ошибка при удалении пользователя: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := repo.Delete(tt.id)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestUserRepository_CheckPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger.Logger = zap.NewNop()
	repo := NewUserRepository(db)

	t.Run("Invalid password", func(t *testing.T) {
		user := entity.User{
			ID:       1,
			Password: "$2a$10$validhash", // Пример хеша
		}
		user.Password = "correctpass"

		mock.ExpectQuery("SELECT id, name, email, password, role FROM users WHERE id =").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role"}).
				AddRow(user.ID, "test", "test@test.com", user.Password, "user"))

		valid := repo.CheckPassword(1, "wrongpass")
		assert.False(t, valid)
	})

	t.Run("User not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name, email, password, role FROM users WHERE id =").
			WithArgs(999).
			WillReturnError(sql.ErrNoRows)

		valid := repo.CheckPassword(999, "anypass")
		assert.False(t, valid)
	})
}
