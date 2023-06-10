package board

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rakateja/milo/twirp-rpc-examples/card/apierror"
	"github.com/rakateja/milo/twirp-rpc-examples/card/database"
)

type SQLRepository struct {
	db        *database.MySQL
	labelRepo LabelRepository
}

const (
	insertBoardQuery = `
		INSERT INTO board (
			entity_id,
			code,
			title,
			created_at,
			updated_at,
			deleted_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`
	updateBoardQuery = `
		UPDATE board SET
			code = ?,
			title = ?,
			created_at = ?,
			updated_at = ?,
			deleted_at = ?
		WHERE entity_id = ?
	`
	selectBoardQuery = `
		SELECT 
			b.entity_id,
			b.code,
			b.title,
			b.created_at,
			b.updated_at,
			b.deleted_at
		FROM board b
	`
	countBoardQuery = `
		SELECT 
			COUNT(entity_id)
		FROM board
	`
	insertMemberQuery = `
		INSERT INTO board_member (
			entity_id,
			board_id,
			user_id,
			created_at,
			updated_at,
			deleted_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`
	updateMemberQuery = `
		UPDATE board_member SET
			board_id = ?,
			user_id = ?,
			created_at = ?,
			updated_at = ?,
			deleted_at = ?
		WHERE entity_id = ?
	`
	deleteMemberQuery = `DELETE FROM board_member`
	selectMemberQuery = `
		SELECT 
			entity_id,
			board_id,
			user_id,
			created_at,
			updated_at,
			deleted_at
		FROM board_member
	`
	countMemberQuery = `
		SELECT 
			COUNT(entity_id)
		FROM board_member
	`
	selectListQuery = `
		SELECT 
			entity_id,
			board_id,
			public_id,
			title,
			position,
			created_at,
			updated_at,
			deleted_at
		FROM board_list
	`
	insertListQuery = `
		INSERT INTO board_list (
			entity_id,
			board_id,
			public_id,
			title,
			position,
			created_at,
			updated_at,
			deleted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	updateListQuery = `
		UPDATE board_list SET 
			board_id = ?,
			public_id = ?, 
			title = ?,
			position = ?,
			created_at = ?,
			updated_at = ?, 
			deleted_at = ?, 
		WHERE entity_id = ?
	`
	countListQuery = `
		SELECT 
			COUNT(entity_id)
		FROM board_list
	`
	deleteListQuery = `
		DELETE FROM board_list
	`
)

func NewSQLRepository(db *database.MySQL) Repository {
	return &SQLRepository{db: db, labelRepo: NewLabelSQLRepository(db)}
}

func (repo *SQLRepository) Store(ctx context.Context, entity *Board) error {
	exist, err := repo.existByID(entity.ID)
	if err != nil {
		return errors.Wrap(err, "exist by id")
	}
	var memberIDs []string
	for _, m := range entity.Members {
		memberIDs = append(memberIDs, m.ID)
	}
	err = repo.db.WithTransaction(func(tx *sqlx.Tx) error {
		if exist {
			err = repo.deleteMember(tx, entity.ID)
			if err != nil {
				return errors.WithStack(err)
			}
			for _, m := range entity.Members {
				err = repo.insertMember(tx, m)
				if err != nil {
					return errors.WithStack(err)
				}
			}
			err = repo.deleteLists(tx, entity.ID)
			if err != nil {
				return errors.WithStack(err)
			}
			for _, l := range entity.Lists {
				err = repo.insertBoardList(tx, l)
				if err != nil {
					return errors.WithStack(err)
				}
			}
			return repo.update(tx, entity)
		}
		err = repo.insert(tx, entity)
		if err != nil {
			return errors.WithStack(err)
		}
		for _, m := range entity.Members {
			err = repo.insertMember(tx, m)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		for _, l := range entity.Lists {
			err = repo.insertBoardList(tx, l)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "store board")
	}
	return nil
}

func (repo *SQLRepository) StoreMember(ctx context.Context, entity BoardMember) error {
	res, err := repo.existMemberByIDs([]string{entity.ID})
	if err != nil {
		return errors.WithMessage(err, "exist by id")
	}
	exist, _ := res[entity.ID]
	err = repo.db.WithTransaction(func(tx *sqlx.Tx) error {
		if exist {
			return repo.updateMember(tx, entity)
		}
		return repo.insertMember(tx, entity)
	})
	if err != nil {
		return errors.WithMessage(err, "store member")
	}
	return nil
}

func (repo *SQLRepository) StoreList(ctx context.Context, entity BoardList) error {
	exist, err := repo.existListByID(entity.ID)
	if err != nil {
		return errors.WithStack(err)
	}
	err = repo.db.WithTransaction(func(tx *sqlx.Tx) error {
		if exist {
			return repo.updateBoardList(tx, entity)
		}
		return repo.insertBoardList(tx, entity)
	})
	return errors.WithStack(err)
}

func (repo *SQLRepository) ResolveByID(ctx context.Context, id string) (*Board, error) {
	var res Board
	err := repo.db.Get(&res, selectBoardQuery+" WHERE b.entity_id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apierror.WithDesc(ErrorCodeEntityNotFound, "board couldn't be found")
		}
		return nil, errors.WithMessage(err, "select board by id")
	}
	var members []BoardMember
	var lists []BoardList
	err = repo.db.Select(&members, selectMemberQuery+" WHERE board_id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, "select board member by board id")
	}
	err = repo.db.Select(&lists, selectListQuery+" WHERE board_id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, "select board list by board id")
	}
	labels, err := repo.labelRepo.ResolveAllByBoardID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "resolve all label by board id")
	}
	res.Members = members
	res.Lists = lists
	res.Labels = labels
	return &res, nil
}

func (repo *SQLRepository) ResolveAllByFilter(ctx context.Context, filter Filter) ([]Board, error) {
	if filter.IsEmpty() {
		return nil, nil
	}
	values := make(map[string]interface{}, 0)
	params := make([]string, 0)
	var innerJoinQuery string
	if filter.UserID != nil {
		innerJoinQuery = "INNER JOIN board_member m ON m.board_id = b.entity_id"
		params = append(params, "m.user_id = :user_id")
		values["user_id"] = *filter.UserID
	}
	whereClauseQuery := ""
	if len(params) > 0 {
		whereClauseQuery = "WHERE " + strings.Join(params, " AND ")
	}
	query, args, err := repo.db.In(selectBoardQuery+" "+innerJoinQuery+" "+whereClauseQuery, values)
	if err != nil {
		err = errors.WithMessage(err, "build where clause sql query")
		return nil, err
	}
	var res []Board
	err = repo.db.Select(&res, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "execute select query")
	}
	return res, nil
}

func (repo *SQLRepository) ResolveAll(ctx context.Context, offset, limit int) ([]Board, error) {
	var res []Board
	err := repo.db.Select(&res, selectBoardQuery+" LIMIT ? OFFSET ?", limit, offset)
	return res, err
}

func (repo *SQLRepository) ResolveTotal(ctx context.Context) (int, error) {
	var total int
	err := repo.db.Get(&total, countBoardQuery)
	return total, err
}

func (repo *SQLRepository) ResolveListByID(ctx context.Context, id string) (BoardList, error) {
	var list BoardList
	err := repo.db.Get(&list, selectListQuery+" WHERE entity_id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return list, apierror.WithDesc(ErrorCodeEntityNotFound, "board list not found")
		}
		return list, errors.WithStack(err)
	}
	return list, nil
}

func (repo *SQLRepository) existByID(id string) (bool, error) {
	var total int
	err := repo.db.Get(&total, countBoardQuery+" WHERE entity_id = ?", id)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return total > 0, nil
}

func (repo *SQLRepository) insert(tx *sqlx.Tx, entity *Board) error {
	res, err := tx.Exec(insertBoardQuery,
		entity.ID,
		entity.Code,
		entity.Title,
		entity.CreatedAt,
		entity.UpdatedAt,
		entity.DeletedAt,
	)
	if err != nil {
		return errors.WithMessage(err, "insert board")
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.WithMessage(err, "rows affected")
	}
	if rowsAffected <= 0 {
		return errors.New("no rows affected")
	}
	return nil
}

func (repo *SQLRepository) update(tx *sqlx.Tx, entity *Board) error {
	res, err := tx.Exec(updateBoardQuery,
		entity.Code,
		entity.Title,
		entity.CreatedAt,
		entity.UpdatedAt,
		entity.DeletedAt,
		entity.ID,
	)
	if err != nil {
		return errors.WithMessage(err, "update board")
	}
	_, err = res.RowsAffected()
	if err != nil {
		return errors.WithMessage(err, "check rows affected")
	}
	return nil
}

func (repo *SQLRepository) existMemberByIDs(ids []string) (map[string]bool, error) {
	sqlQuery := `
		SELECT
			entity_id,
			COUNT(entity_id)
		FROM board_member
		WHERE entity_id IN (:id) GROUP BY entity_id
	`
	res := make(map[string]bool, 0)
	query, args, err := repo.db.In(sqlQuery, map[string]interface{}{"id": ids})
	if err != nil {
		return res, errors.WithStack(err)
	}
	query = repo.db.Rebind(query)
	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return res, errors.WithStack(err)
	}
	for rows.Next() {
		var id string
		var total int
		err = rows.Scan(&id, &total)
		if err != nil {
			return res, errors.Wrap(err, "scan sql rows")
		}
		if total > 0 {
			res[id] = true
			continue
		}
		res[id] = false
	}
	return res, nil
}

func (repo *SQLRepository) insertMember(tx *sqlx.Tx, entity BoardMember) error {
	res, err := tx.Exec(insertMemberQuery,
		entity.ID,
		entity.BoardID,
		entity.UserID,
		entity.CreatedAt,
		entity.UpdatedAt,
		entity.DeletedAt,
	)
	if err != nil {
		return errors.WithMessage(err, "insert member")
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.WithMessage(err, "check rows affected")
	}
	if rowsAffected <= 0 {
		return errors.New("insert member fails")
	}
	return nil
}

func (repo *SQLRepository) deleteMember(tx *sqlx.Tx, boardID string) error {
	_, err := tx.Exec(deleteMemberQuery+" WHERE board_id = ?", boardID)
	if err != nil {
		return errors.Wrap(err, "delete members")
	}
	return nil
}

func (repo *SQLRepository) deleteLists(tx *sqlx.Tx, boardID string) error {
	_, err := tx.Exec(deleteListQuery+" WHERE board_id = ?", boardID)
	if err != nil {
		return errors.Wrap(err, "delete lists")
	}
	return nil
}

func (repo *SQLRepository) updateMember(tx *sqlx.Tx, entity BoardMember) error {
	res, err := tx.Exec(updateMemberQuery,
		entity.BoardID,
		entity.UserID,
		entity.CreatedAt,
		entity.UpdatedAt,
		entity.DeletedAt,
		entity.ID,
	)
	if err != nil {
		return errors.WithMessage(err, "insert member")
	}
	_, err = res.RowsAffected()
	if err != nil {
		return errors.WithMessage(err, "check rows affected")
	}
	return nil
}

func (repo *SQLRepository) existListByID(id string) (bool, error) {
	var total int
	err := repo.db.Get(&total, countListQuery+" WHERE entity_id = ?", id)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return total > 0, nil
}

func (repo *SQLRepository) insertBoardList(tx *sqlx.Tx, entity BoardList) error {
	res, err := tx.Exec(insertListQuery,
		entity.ID,
		entity.BoardID,
		entity.PublicID,
		entity.Title,
		entity.Position,
		entity.CreatedAt,
		entity.UpdatedAt,
		entity.DeletedAt,
	)
	if err != nil {
		return errors.WithStack(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.WithStack(err)
	}
	if rowsAffected <= 0 {
		return errors.New("no rows affected")
	}
	return nil
}

func (repo *SQLRepository) updateBoardList(tx *sqlx.Tx, entity BoardList) error {
	res, err := tx.Exec(updateListQuery,
		entity.BoardID,
		entity.PublicID,
		entity.Title,
		entity.Position,
		entity.CreatedAt,
		entity.UpdatedAt,
		entity.DeletedAt,
		entity.ID,
	)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
