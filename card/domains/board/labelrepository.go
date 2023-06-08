package board

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rakateja/milo/twirp-rpc-examples/card/database"
)

type LabelRepository interface {
	Store(ctx context.Context, entity *Label) error
	ResolveByID(ctx context.Context, id string) (*Label, error)
	ExistByID(ctx context.Context, id string) (bool, error)
	ResolveAllByBoardID(ctx context.Context, boardID string) ([]Label, error)
	ResolveBySlug(ctx context.Context, slug string) (*Label, error)
}

type LabelSQLRepository struct {
	db *database.MySQL
}

func NewLabelSQLRepository(db *database.MySQL) LabelRepository {
	return &LabelSQLRepository{db: db}
}

const (
	insertLabelQuery = `
		INSERT INTO label (
			entity_id,
			board_id,
			slug,
			title,
			color,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	updateLabelQuery = `
		UPDATE label SET
			board_id = ?,
			slug = ?,
			title = ?,
			color = ?,
			created_at = ?,
			updated_at = ?
		WHERE entity_id = ?
	`
	selectLabelQuery = `
		SELECT
			entity_id,
			board_id,
			slug,
			title,
			color,
			created_at,
			updated_at
		FROM label	
	`
	countLabelQuery = `
		SELECT COUNT(entity_id) FROM label
	`
)

func (repo *LabelSQLRepository) Store(ctx context.Context, entity *Label) error {
	exist, err := repo.existByID(ctx, entity.ID)
	if err != nil {
		return errors.Wrap(err, "exist label by id")
	}
	return repo.db.WithTransaction(func(tx *sqlx.Tx) error {
		if exist {
			return repo.update(tx, entity)
		}
		return repo.insert(tx, entity)
	})
}

func (repo *LabelSQLRepository) ResolveByID(ctx context.Context, id string) (*Label, error) {
	var res Label
	err := repo.db.Get(&res, selectLabelQuery+" WHERE entity_id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, "select label by id")
	}
	return &res, nil
}

func (repo *LabelSQLRepository) ExistByID(ctx context.Context, id string) (bool, error) {
	return repo.existByID(ctx, id)
}

func (repo *LabelSQLRepository) ResolveAllByBoardID(ctx context.Context, boardID string) ([]Label, error) {
	var res []Label
	err := repo.db.Select(&res, selectLabelQuery+" WHERE board_id = ?", boardID)
	if err != nil {
		return nil, errors.Wrap(err, "select label by board id")
	}
	return res, nil
}

func (repo *LabelSQLRepository) ResolveBySlug(ctx context.Context, slug string) (res *Label, err error) {
	err = repo.db.Get(&res, selectLabelQuery+" WHERE slug = ?", slug)
	if err != nil {
		err = errors.Wrap(err, "select label by slug")
		return
	}
	return res, nil
}

func (repo *LabelSQLRepository) existByID(ctx context.Context, id string) (bool, error) {
	var total int
	err := repo.db.Get(&total, countLabelQuery+" WHERE entity_id = ?", id)
	if err != nil {
		return false, errors.Wrap(err, "count label by id")
	}
	return total > 0, nil
}

func (repo *LabelSQLRepository) insert(tx *sqlx.Tx, entity *Label) error {
	res, err := tx.Exec(insertLabelQuery,
		entity.ID,
		entity.BoardID,
		entity.Slug,
		entity.Title,
		entity.Color,
		entity.CreatedAt,
		entity.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "insert label")
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "checking rows affected")
	}
	if rowsAffected <= 0 {
		return errors.Wrap(err, "zero affected rows")
	}
	return nil
}

func (repo *LabelSQLRepository) update(tx *sqlx.Tx, entity *Label) error {
	_, err := tx.Exec(updateLabelQuery,
		entity.BoardID,
		entity.Slug,
		entity.Title,
		entity.Color,
		entity.CreatedAt,
		entity.UpdatedAt,
		entity.ID,
	)
	if err != nil {
		return errors.Wrap(err, "update label")
	}
	return nil
}
