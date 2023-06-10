package card

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rakateja/milo/twirp-rpc-examples/card/apierror"
	"github.com/rakateja/milo/twirp-rpc-examples/card/database"
)

type SQLRepository struct {
	db *database.MySQL
}

const (
	insertCardQuery = `
		INSERT INTO card (
			entity_id,
			list_id,
			board_id,
			public_id,
			title,
			description,
			due_date_from,
			due_date_until,
			due_date_completed_at,
			created_at,
			updated_at,
			deleted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	updateCardQuery = `
		UPDATE card SET
			list_id = ?,
			board_id = ?,
			public_id = ?,
			title = ?,
			description = ?,
			due_date_from = ?,
			due_date_until = ?,
			due_date_completed_at = ?,
			updated_at = ?,
			deleted_at = ?
		WHERE entity_id = ?
	`
	selectCardIDQuery = `
		SELECT
			c.entity_id
		FROM card c
	`
	selectCardQuery = `
		SELECT
			c.entity_id,
			c.list_id,
			c.board_id,
			c.public_id,
			c.title,
			c.description,
			c.due_date_from,
			c.due_date_until,
			c.due_date_completed_at,
			c.created_at,
			c.updated_at,
			c.deleted_at
		FROM card c
	`
	countCardQuery = `
		SELECT
			COUNT(entity_id)
		FROM card c
	`
	insertMemberQuery = `
		INSERT INTO card_member (
			entity_id,
			card_id,
			user_id,
			created_at
		) VALUES (?, ?, ?, ?)
	`
	deleteMemberQuery = `
		DELETE FROM card_member
	`
	selectMemberQuery = `
		SELECT
			entity_id,
			card_id,
			user_id,
			created_at
		FROM card_member
	`
	countMemberQuery = `
		SELECT 
			COUNT(entity_id)
		FROM card_member
	`
	insertAttachmentQuery = `
		INSERT INTO card_attachment (
			entity_id,
			card_id,
			link_name,
			file_type,
			file_url,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	deleteAttachmentQuery = `
		DELETE FROM card_attachment
	`
	selectAttachmentQuery = `
		SELECT 
			entity_id,
			card_id,
			link_name,
			file_type,
			file_url,
			created_at,
			updated_at
		FROM card_attachment
	`
	countAttachmentQuery = `
		SELECT 
			COUNT(entity_id)
		FROM card_attachment
	`
	selectLabelQuery = `
		SELECT
			entity_id,
			card_id,
			label_id,
			created_at
		FROM card_label
	`
	insertLabelQuery = `
		INSERT INTO card_label (entity_id, card_id, label_id, created_at)
		VALUES (?, ?, ?, ?)
	`
	deleteLabelQuery = `
		DELETE FROM card_label
	`
)

func NewSQLRepository(db *database.MySQL) Repository {
	return &SQLRepository{db: db}
}

func (repo *SQLRepository) Store(ctx context.Context, entity *Card) error {
	exist, err := repo.existByID(ctx, entity.ID)
	if err != nil {
		return errors.Wrap(err, "exist by id")
	}
	err = repo.db.WithTransaction(func(tx *sqlx.Tx) error {
		if exist {
			return repo.update(tx, entity)
		}
		return repo.insert(tx, entity)
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (repo *SQLRepository) StoreLabels(ctx context.Context, cardID string, labels []Label) error {
	err := repo.db.WithTransaction(func(tx *sqlx.Tx) error {
		err := repo.deleteLabelsByCardID(tx, cardID)
		if err != nil {
			return errors.WithStack(err)
		}
		err = repo.insertLabels(tx, labels)
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (repo *SQLRepository) ResolveByID(ctx context.Context, id string) (*Card, error) {
	log.Println("ResolveByID() is invoked")
	var result Card
	err := repo.db.Get(&result, selectCardQuery+" WHERE c.entity_id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apierror.WithDesc(ErrorCodeEntityNotFound, "card couldn't be found")
		}
		return nil, errors.Wrap(err, "select card by id")
	}
	members, err := repo.resolveMembersByCardID(ctx, []string{result.ID})
	if err != nil {
		return nil, errors.Wrap(err, "resolve member by card ids")
	}
	attachments, err := repo.resolveAttachmentsByCardID(ctx, []string{result.ID})
	if err != nil {
		return nil, errors.Wrap(err, "resolve attachment by card ids")
	}
	labels, err := repo.selectLabelsByCardID(id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	result.Members = members
	result.Attachments = attachments
	result.Labels = labels
	return &result, nil
}

func (repo *SQLRepository) ResolveIDsByFilter(ctx context.Context, filter Filter) ([]string, error) {
	if filter.IsEmpty() {
		return make([]string, 0), nil
	}
	whereClauseQuery, joinQuery, values := repo.buildQueryWithFilter(filter)
	query, args, err := repo.db.In(selectCardIDQuery+" "+joinQuery+" "+whereClauseQuery, values)
	if err != nil {
		return make([]string, 0), errors.WithStack(err)
	}
	query = repo.db.Rebind(query)
	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return make([]string, 0), errors.WithStack(err)
	}
	var cardIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return make([]string, 0), err
		}
		cardIDs = append(cardIDs, id)
	}
	return cardIDs, nil
}

func (repo *SQLRepository) ResolveAllByFilter(ctx context.Context, filter Filter) ([]Card, error) {
	if filter.IsEmpty() {
		return nil, nil
	}
	whereClauseQuery, joinQuery, values := repo.buildQueryWithFilter(filter)
	log.Printf("query: %s", joinQuery+whereClauseQuery)
	query, args, err := repo.db.In(selectCardQuery+" "+joinQuery+" "+whereClauseQuery, values)
	if err != nil {
		err = errors.WithStack(err)
		return nil, err
	}
	query = repo.db.Rebind(query)
	var res []Card
	err = repo.db.Select(&res, query, args...)
	if err != nil {
		err = errors.WithMessage(err, "select card with filter")
		return nil, err
	}
	var cardIDs []string
	for _, entity := range res {
		cardIDs = append(cardIDs, entity.ID)
	}
	members, err := repo.resolveMembersByCardID(ctx, cardIDs)
	if err != nil {
		err = errors.Wrap(err, "resolve members by card id")
		return nil, err
	}
	attachments, err := repo.resolveAttachmentsByCardID(ctx, cardIDs)
	if err != nil {
		err = errors.Wrap(err, "resolve attacment by card id")
		return nil, err
	}
	membersMap := make(map[string][]Member, 0)
	attachmentsMap := make(map[string][]Attachment, 0)
	for _, m := range members {
		membersMap[m.CardID] = append(membersMap[m.CardID], m)
	}
	for _, a := range attachments {
		attachmentsMap[a.CardID] = append(attachmentsMap[a.CardID], a)
	}
	var result []Card
	for _, cardEntity := range res {
		cardEntity.Attachments = attachmentsMap[cardEntity.ID]
		cardEntity.Members = membersMap[cardEntity.ID]
		result = append(result, cardEntity)
	}
	return result, nil
}

func (repo *SQLRepository) CountByFilter(ctx context.Context, filter Filter) (int, error) {
	whereClauseQuery, joinQuery, values := repo.buildQueryWithFilter(filter)
	query, args, err := repo.db.In(countCardQuery+" "+joinQuery+" "+whereClauseQuery, values)
	if err != nil {
		err = errors.WithStack(err)
		return 0, err
	}
	query = repo.db.Rebind(query)
	var total int
	err = repo.db.Get(&total, query, args...)
	if err != nil {
		return total, errors.Wrap(err, "count rows by filter")
	}
	return total, nil
}

func (repo *SQLRepository) buildQueryWithFilter(filter Filter) (whereClauseQuery string, joinQuery string, values map[string]interface{}) {
	values = make(map[string]interface{}, 0)
	params := make([]string, 0)
	if len(filter.IDs) > 0 {
		params = append(params, "c.entity_id IN (:entity_ids)")
		values["entity_ids"] = filter.IDs
	}
	if len(filter.PublicIDs) > 0 {
		params = append(params, "c.public_id IN (:public_ids)")
		values["public_ids"] = filter.PublicIDs
	}
	if len(filter.ListIDs) > 0 {
		params = append(params, "c.list_id IN (:list_ids)")
		values["list_ids"] = filter.ListIDs
	}
	if len(filter.BoardIDs) > 0 {
		params = append(params, "c.board_id IN (:board_ids)")
		values["board_ids"] = filter.BoardIDs
	}
	if len(filter.UserIDs) > 0 {
		joinQuery = "INNER JOIN card_member m ON m.card_id = c.entity_id"
		params = append(params, "m.user_id IN (:user_ids)")
		values["user_ids"] = filter.UserIDs
	}
	whereClauseQuery = "WHERE " + strings.Join(params, " AND ")
	return whereClauseQuery, joinQuery, values
}

func (repo *SQLRepository) resolveMembersByCardID(ctx context.Context, cardIDs []string) (res []Member, err error) {
	query, args, err := repo.db.In(selectMemberQuery+" WHERE card_id IN (:card_id)", map[string]interface{}{
		"card_id": cardIDs,
	})
	if err != nil {
		return
	}
	query = repo.db.Rebind(query)
	err = repo.db.Select(&res, query, args...)
	if err != nil {
		err = errors.Wrap(err, "resolve member by card id")
		return
	}
	return res, nil
}

func (repo *SQLRepository) resolveAttachmentsByCardID(ctx context.Context, cardIDs []string) (res []Attachment, err error) {
	query, args, err := repo.db.In(selectAttachmentQuery+" WHERE card_id IN (:card_id)", map[string]interface{}{
		"card_id": cardIDs,
	})
	if err != nil {
		return
	}
	query = repo.db.Rebind(query)
	err = repo.db.Select(&res, query, args...)
	if err != nil {
		err = errors.Wrap(err, "resolve attachment by card id")
		return
	}
	return res, nil
}

func (repo *SQLRepository) existByID(ctx context.Context, id string) (bool, error) {
	var total int
	err := repo.db.Get(&total, countCardQuery+" WHERE entity_id = ?", id)
	if err != nil {
		return false, errors.Wrap(err, "execute count card query")
	}
	return total > 0, nil
}

func (repo *SQLRepository) insert(tx *sqlx.Tx, entity *Card) error {
	res, err := tx.Exec(insertCardQuery,
		entity.ID,
		entity.ListID,
		entity.BoardID,
		entity.PublicID,
		entity.Title,
		entity.Description,
		entity.DueDateFrom,
		entity.DueDateUntil,
		entity.DueDateCompletedAt,
		entity.CreatedAt,
		entity.UpdatedAt,
		entity.DeletedAt,
	)
	if err != nil {
		return errors.Wrap(err, "insert card")
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "checking rows affected")
	}
	if rowsAffected <= 0 {
		return errors.New("no rows affected")
	}
	err = repo.insertMembers(tx, entity.Members)
	if err != nil {
		return errors.Wrap(err, "insert members")
	}
	err = repo.insertAttachments(tx, entity.Attachments)
	if err != nil {
		return errors.Wrap(err, "insert attachments")
	}
	return nil
}

func (repo *SQLRepository) update(tx *sqlx.Tx, entity *Card) error {
	res, err := tx.Exec(updateCardQuery,
		entity.ListID,
		entity.BoardID,
		entity.PublicID,
		entity.Title,
		entity.Description,
		entity.DueDateFrom,
		entity.DueDateUntil,
		entity.DueDateCompletedAt,
		entity.UpdatedAt,
		entity.DeletedAt,
		entity.ID,
	)
	if err != nil {
		return errors.Wrap(err, "insert card")
	}
	_, err = res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "checking rows affected")
	}
	// update members
	err = repo.deleteMemberByCardID(tx, entity.ID)
	if err != nil {
		return errors.Wrap(err, "delete members by card id")
	}
	err = repo.insertMembers(tx, entity.Members)
	if err != nil {
		return errors.Wrap(err, "insert members")
	}
	// update attachments
	err = repo.deleteAttachmentsByCardID(tx, entity.ID)
	if err != nil {
		return errors.Wrap(err, "delete attachment")
	}
	err = repo.insertAttachments(tx, entity.Attachments)
	if err != nil {
		return errors.Wrap(err, "insert attachment")
	}
	return nil
}

func (repo *SQLRepository) insertMembers(tx *sqlx.Tx, memberList []Member) error {
	for _, m := range memberList {
		res, err := tx.Exec(insertMemberQuery,
			m.ID,
			m.CardID,
			m.UserID,
			m.CreatedAt,
		)
		if err != nil {
			return errors.Wrap(err, "insert member")
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return errors.Wrap(err, "checking rows affected")
		}
		if rowsAffected <= 0 {
			return errors.Wrap(err, "no rows affected")
		}
	}
	return nil
}

func (repo *SQLRepository) deleteMemberByCardID(tx *sqlx.Tx, id string) error {
	_, err := tx.Exec(deleteMemberQuery+" WHERE card_id = ?", id)
	if err != nil {
		return errors.Wrap(err, "delete member by card id")
	}
	return nil
}

func (repo *SQLRepository) insertAttachments(tx *sqlx.Tx, attachments []Attachment) error {
	for _, a := range attachments {
		res, err := tx.Exec(insertAttachmentQuery,
			a.ID,
			a.CardID,
			a.LinkName,
			a.FileType,
			a.FileURL,
			a.CreatedAt,
			a.UpdatedAt,
		)
		if err != nil {
			return errors.Wrap(err, "insert attachment")
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return errors.Wrap(err, "update attachment")
		}
		if rowsAffected <= 0 {
			return errors.New("No affected rows")
		}
	}
	return nil
}

func (repo *SQLRepository) deleteAttachmentsByCardID(tx *sqlx.Tx, cardID string) error {
	_, err := tx.Exec(deleteAttachmentQuery+" WHERE card_id = ?", cardID)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (repo *SQLRepository) selectLabelsByCardID(cardID string) ([]Label, error) {
	var res []Label
	err := repo.db.Select(&res, selectLabelQuery+" WHERE card_id = ?", cardID)
	if err != nil {
		return res, errors.Wrap(err, "select card label with card id")
	}
	return res, nil
}

func (repo *SQLRepository) deleteLabelsByCardID(tx *sqlx.Tx, cardID string) error {
	_, err := tx.Exec(deleteLabelQuery+" WHERE card_id = ?", cardID)
	if err != nil {
		return errors.Wrap(err, "delete card labels row by card id")
	}
	return nil
}

func (repo *SQLRepository) insertLabels(tx *sqlx.Tx, labels []Label) error {
	for _, label := range labels {
		_, err := tx.Exec(
			insertLabelQuery,
			label.ID,
			label.CardID,
			label.LabelID,
			label.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}
