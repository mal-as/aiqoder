package repos

import "github.com/mal-as/aiqoder/pkg/pg/transaction"

type Repository struct {
	db *transaction.SQLManager
}

func NewRepository(db *transaction.SQLManager) *Repository {
	return &Repository{db: db}
}
