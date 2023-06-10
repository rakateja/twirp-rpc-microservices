package card

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rakateja/milo/twirp-rpc-examples/card/apierror"
)

const (
	ErrorCodeEntityNotFound = "EntityNotFound"
	ErrorCodeInvalidInput   = "InvalidInput"
	ErrorCodeAlreadyExist   = "AlreadyExist"
)

type Card struct {
	ID                 string       `json:"entity_id" db:"entity_id"`
	ListID             string       `json:"list_id" db:"list_id"`
	BoardID            string       `json:"board_id" db:"board_id"`
	PublicID           string       `json:"public_id" db:"public_id"`
	Title              string       `json:"title" db:"title"`
	Description        string       `json:"description" db:"description"`
	DueDateFrom        *time.Time   `json:"due_date_from" db:"due_date_from"`
	DueDateUntil       *time.Time   `json:"due_date_until" db:"due_date_until"`
	DueDateCompletedAt *time.Time   `json:"due_date_completed_at" db:"due_date_completed_at"`
	Members            []Member     `json:"members"`
	Attachments        []Attachment `json:"attachments"`
	Labels             []Label      `json:"labels"`
	CreatedAt          time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt          *time.Time   `json:"deleted_at" db:"deleted_at"`
}

func (c *Card) MoveList(listID string) error {
	c.ListID = listID
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Card) AddAttachment(fileURL, fileType, linkName string) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return errors.WithStack(err)
	}
	now := time.Now()
	c.Attachments = append(c.Attachments, Attachment{
		ID:        id.String(),
		CardID:    c.ID,
		LinkName:  linkName,
		FileURL:   fileURL,
		FileType:  fileType,
		CreatedAt: now,
		UpdatedAt: now,
	})
	c.UpdatedAt = now
	return nil
}

func (c *Card) DeleteAttachment(attachmentID string) error {
	updatedAttachments := make([]Attachment, 0)
	for _, t := range c.Attachments {
		fmt.Printf("%s %s\n", t.ID, attachmentID)
		if strings.Trim(t.ID, " ") == strings.Trim(attachmentID, " ") {
			continue
		}
		updatedAttachments = append(updatedAttachments, t)
	}
	if len(c.Attachments) == len(updatedAttachments) {
		return apierror.WithDesc(ErrorCodeEntityNotFound, "attachment not found")
	}
	c.Attachments = updatedAttachments
	return nil
}

func (c *Card) UpdateLinkName(attachmentID string, linkName string) error {
	if linkName == "" {
		return apierror.WithDesc(ErrorCodeInvalidInput, "linkName is mandatory")
	}
	updatedAttachments := make([]Attachment, 0)
	for _, t := range c.Attachments {
		if t.ID != attachmentID {
			updatedAttachments = append(updatedAttachments, t)
			continue
		}
		t.LinkName = linkName
		updatedAttachments = append(updatedAttachments, t)
	}
	c.Attachments = updatedAttachments
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Card) Update(input CardInput) error {
	validate := validator.New()
	err := validate.Struct(input)
	if err != nil {
		return errors.Wrap(err, "validate card input")
	}
	c.Title = input.Title
	c.Description = input.Description
	c.ListID = input.ListID
	c.BoardID = input.BoardID
	if input.DueDateFrom != nil && input.DueDateUntil != nil {
		dueDateFrom, err := time.Parse(time.RFC3339, *input.DueDateFrom)
		if err != nil {
			return errors.Wrap(err, "parse due date from")
		}
		dueDateUntil, err := time.Parse(time.RFC3339, *input.DueDateUntil)
		if err != nil {
			return errors.Wrap(err, "parse due date until")
		}
		c.DueDateFrom = &dueDateFrom
		c.DueDateUntil = &dueDateUntil
	}
	now := time.Now()
	if c.DueDateCompletedAt == nil && input.DueDateIsCompleted {
		c.DueDateCompletedAt = &now
	}
	c.UpdatedAt = now
	return nil
}

func (c *Card) UpdateMembers(members []MemberInput) error {
	currentMemberMap := make(map[string]Member, 0)
	for _, m := range c.Members {
		currentMemberMap[m.UserID] = m
	}
	memberInputMap := make(map[string]MemberInput, 0)
	for _, m := range members {
		memberInputMap[m.UserID] = m
	}
	newMembers := make([]string, 0)
	deletedMembers := make([]string, 0)
	for _, m := range c.Members {
		_, exist := memberInputMap[m.UserID]
		if !exist {
			deletedMembers = append(deletedMembers, m.UserID)
		}
	}
	updatedMembers := make([]Member, 0)
	now := time.Now()
	for _, m := range members {
		memberEntity, exist := currentMemberMap[m.UserID]
		if exist {
			updatedMembers = append(updatedMembers, memberEntity)
			continue
		}
		id, err := uuid.NewUUID()
		if err != nil {
			return err
		}
		newMembers = append(newMembers, m.UserID)
		updatedMembers = append(updatedMembers, Member{
			ID:        id.String(),
			CardID:    c.ID,
			UserID:    m.UserID,
			CreatedAt: now,
		})
	}
	c.Members = updatedMembers
	c.UpdatedAt = now
	return nil
}

func (c *Card) AddLabel(labelID string) error {
	for _, label := range c.Labels {
		if label.LabelID == labelID {
			return apierror.New(ErrorCodeAlreadyExist)
		}
	}
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	c.Labels = append(c.Labels, Label{
		ID:        id.String(),
		CardID:    c.ID,
		LabelID:   labelID,
		CreatedAt: time.Now(),
	})
	return nil
}

func (c *Card) RemoveLabel(labelID string) {
	var updatedLabels []Label
	for _, label := range c.Labels {
		if label.LabelID == labelID {
			continue
		}
		updatedLabels = append(updatedLabels, label)
	}
}

type Member struct {
	ID        string    `json:"entity_id" db:"entity_id"`
	CardID    string    `json:"card_id" db:"card_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Attachment struct {
	ID        string    `json:"entity_id" db:"entity_id"`
	CardID    string    `json:"card_id" db:"card_id"`
	LinkName  string    `json:"link_name" db:"link_name"`
	FileType  string    `json:"file_type" db:"file_type"`
	FileURL   string    `json:"file_url" db:"file_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Label struct {
	ID        string    `json:"entity_id" db:"entity_id"`
	CardID    string    `json:"card_id" db:"card_id"`
	LabelID   string    `json:"label_id" db:"label_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CardInput struct {
	ListID             string            `json:"list_id" validate:"required"`
	BoardID            string            `json:"board_id" validate:"required"`
	Title              string            `json:"title" validate:"required"`
	Description        string            `json:"description"`
	DueDateFrom        *string           `json:"due_date_from"`
	DueDateUntil       *string           `json:"due_date_until"`
	DueDateIsCompleted bool              `json:"due_date_is_completed"`
	Members            []MemberInput     `json:"members"`
	Attachments        []AttachmentInput `json:"attachments"`
}

func (input CardInput) ToEntity() (*Card, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	now := time.Now()
	var members []Member
	var attachments []Attachment
	for _, m := range input.Members {
		memberID, err := uuid.NewUUID()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		members = append(members, Member{
			ID:        memberID.String(),
			CardID:    id.String(),
			UserID:    m.UserID,
			CreatedAt: now,
		})
	}
	for _, a := range input.Attachments {
		attachmentID, err := uuid.NewUUID()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		attachments = append(attachments, Attachment{
			ID:        attachmentID.String(),
			CardID:    id.String(),
			LinkName:  a.LinkName,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	return &Card{
		ID:                 id.String(),
		BoardID:            input.BoardID,
		ListID:             input.ListID,
		Title:              input.Title,
		Description:        input.Description,
		DueDateFrom:        nil,
		DueDateUntil:       nil,
		DueDateCompletedAt: nil,
		Members:            members,
		Attachments:        attachments,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

type MemberInput struct {
	UserID string `json:"user_id" validate:"required"`
}

type MemberListInput struct {
	Members []MemberInput `json:"members"`
}

type AttachmentInput struct {
	LinkName   string `json:"link_name" validate:"required"`
	FileBase64 string `json:"file_base64" validate:"base64"`
}
