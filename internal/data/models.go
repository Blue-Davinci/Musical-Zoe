package data

import (
	"errors"

	"github.com/Blue-Davinci/musical-zoe/internal/database"
)

var (
	ErrGeneralRecordNotFound = errors.New("finance record not found")
	ErrGeneralEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Users  UserModel
	Tokens TokenModel
}

func NewModels(db *database.Queries) Models {
	return Models{
		Users:  UserModel{DB: db},
		Tokens: TokenModel{DB: db},
	}
}
