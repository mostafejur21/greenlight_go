package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// The Models struct wraps the MovieModel. we'll add other models to this,
// like UserModel, PermissionModel ect
type Models struct {
	Movies MovieModel
	Users  UserModel
}

// NewModels will initialize the MovieModel
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Users:  UserModel{DB: db},
	}
}
