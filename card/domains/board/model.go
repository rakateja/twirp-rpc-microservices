package board

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rakateja/milo/twirp-rpc-examples/card/apierror"
)

const (
	ErrorCodeEntityNotFound = "EntityNotFound"
)

type Filter struct {
	PublicIDs []string `json:"public_ids"`
	UserID    *string  `json:"user_id"`
}

func (t Filter) IsEmpty() bool {
	return t.UserID == nil && len(t.PublicIDs) == 0
}

type Board struct {
	ID        string        `json:"entity_id" db:"entity_id"`
	Code      string        `json:"code" db:"code"`
	Title     string        `json:"title" db:"title"`
	Members   []BoardMember `json:"members"`
	Lists     []BoardList   `json:"lists"`
	Labels    []Label       `json:"labels"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time    `json:"deleted_at" db:"deleted_at"`
}

func (b Board) HasAccess(userID string) bool {
	for _, m := range b.Members {
		if m.UserID == userID {
			return true
		}
	}
	return false
}

func (b Board) MemberExist(userID string) bool {
	for _, m := range b.Members {
		if m.UserID == userID {
			return true
		}
	}
	return false
}

func (b *Board) ListExist(listID string) bool {
	for _, l := range b.Lists {
		if l.ID == listID {
			return true
		}
	}
	return false
}

func (b *Board) RemoveMember(userID string) {
	updatedMembers := make([]BoardMember, 0)
	for _, m := range b.Members {
		if m.UserID != userID {
			updatedMembers = append(updatedMembers, m)
		}
	}
	b.Members = updatedMembers
	b.UpdatedAt = time.Now()
}

func (b *Board) Update(input UpdateInput) error {
	mapLists := make(map[string]BoardList, 0)
	for _, list := range b.Lists {
		mapLists[list.ID] = list
	}
	for _, list := range input.Lists {
		boardList, exist := mapLists[list.ID]
		if !exist {
			return apierror.WithDesc(ErrorCodeEntityNotFound, "board list couldn't be found")
		}
		boardList.Title = list.Title
		boardList.Position = list.Position
		boardList.UpdatedAt = time.Now()
		mapLists[list.ID] = boardList
	}
	var updatedList []BoardList
	for _, list := range mapLists {
		updatedList = append(updatedList, list)
	}
	b.Title = input.Title
	b.Lists = updatedList
	b.UpdatedAt = time.Now()
	return nil
}

type MemberInput struct {
	UserID string `json:"user_id" validate:"required"`
}

type MemberListInput struct {
	Members []MemberInput `json:"members"`
}

type ListInput struct {
	Title    string `json:"title" validate:"required"`
	Position int    `json:"position" validate:"required"`
}

type ListUpdateInput struct {
	ID       string `json:"entity_id" validate:"required"`
	Title    string `json:"title" validate:"required"`
	Position int    `json:"position" validate:"required"`
}

type UpdateInput struct {
	Title string            `json:"title" validate:"required"`
	Lists []ListUpdateInput `json:"lists"`
}

type Input struct {
	Title   string        `json:"title" validate:"required"`
	Members []MemberInput `json:"members"`
	Lists   []ListInput   `json:"lists"`
	Labels  []LabelInput  `json:"labels"`
}

func (t Input) ToEntity() (res *Board, err error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return
	}
	now := time.Now()
	var members []BoardMember
	var lists []BoardList
	for _, m := range t.Members {
		boardMember, err := NewBoardMember(id.String(), m.UserID)
		if err != nil {
			return res, err
		}
		members = append(members, boardMember)
	}
	for _, l := range t.Lists {
		boardList, err := NewBoardList(id.String(), l)
		if err != nil {
			return res, err
		}
		lists = append(lists, boardList)
	}
	return &Board{
		ID:        id.String(),
		Code:      "FOOBAR",
		Title:     t.Title,
		Members:   members,
		Lists:     lists,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

type BoardMember struct {
	ID        string     `json:"entity_id" db:"entity_id"`
	BoardID   string     `json:"board_id" db:"board_id"`
	UserID    string     `json:"user_id" db:"user_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}

func NewBoardMember(boardID, userID string) (BoardMember, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return BoardMember{}, err
	}
	now := time.Now()
	return BoardMember{
		ID:        id.String(),
		BoardID:   boardID,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

type BoardList struct {
	ID        string     `json:"entity_id" db:"entity_id"`
	BoardID   string     `json:"board_id" db:"board_id"`
	PublicID  string     `json:"public_id" db:"public_id"`
	Title     string     `json:"title" db:"title"`
	Position  int        `json:"position" db:"position"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}

func NewBoardList(boardID string, input ListInput) (BoardList, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return BoardList{}, err
	}
	now := time.Now()
	return BoardList{
		ID:        id.String(),
		BoardID:   boardID,
		Title:     input.Title,
		Position:  input.Position,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

type Label struct {
	ID        string    `json:"entity_id" db:"entity_id"`
	BoardID   string    `json:"board_id" db:"board_id"`
	Slug      string    `json:"slug" db:"slug"`
	Title     string    `json:"title" db:"title"`
	Color     string    `json:"color" db:"color"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (l *Label) Update(title, color string) {
	l.Title = title
	l.Slug = strings.ReplaceAll(strings.ToLower(l.Title), " ", "_")
	l.Color = color
	l.UpdatedAt = time.Now()
}

type LabelInput struct {
	Title string `json:"title"`
	Color string `json:"color"`
}

func (l LabelInput) ToEntity(boardID string) (*Label, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &Label{
		ID:        id.String(),
		BoardID:   boardID,
		Title:     l.Title,
		Slug:      strings.ReplaceAll(strings.ToLower(l.Title), " ", "_"),
		Color:     l.Color,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
