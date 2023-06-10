package card

import (
	"context"
	"math/rand"

	"github.com/pkg/errors"
	"github.com/rakateja/milo/twirp-rpc-examples/card/apierror"
	"github.com/rakateja/milo/twirp-rpc-examples/card/domains/board"
)

type Service struct {
	repo         Repository
	boardService *board.Service
}

func NewService(repo Repository, boardService *board.Service) *Service {
	return &Service{
		repo:         repo,
		boardService: boardService,
	}
}

func (svc *Service) Create(ctx context.Context, input CardInput) (*Card, error) {
	entity, err := input.ToEntity()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	code, err := svc.generateCode(ctx, 0)
	if err != nil {
		return nil, errors.Wrap(err, "geneate card public id")
	}
	entity.PublicID = code
	err = svc.repo.Store(ctx, entity)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return svc.repo.ResolveByID(ctx, entity.ID)
}

func (svc *Service) Update(ctx context.Context, cardID string, input CardInput) (*Card, error) {
	boardEntity, err := svc.boardService.ResolveByID(ctx, cardID)
	if err != nil {
		return nil, errors.Wrap(err, "resolve board by id")
	}
	if !boardEntity.ListExist(input.ListID) {
		return nil, apierror.WithDesc(ErrorCodeInvalidInput, "invalid list ID")
	}
	entity, err := svc.repo.ResolveByID(ctx, cardID)
	if err != nil {
		return nil, errors.Wrap(err, "resolve card by id")
	}
	if entity.ListID != input.ListID {
	}
	err = entity.Update(input)
	if err != nil {
		return nil, errors.Wrap(err, "update card")
	}
	return entity, nil
}

func (svc *Service) MoveList(ctx context.Context, cardID, listID string) (*Card, error) {
	entity, err := svc.repo.ResolveByID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if err := entity.MoveList(listID); err != nil {
		return nil, err
	}
	if err = svc.repo.Store(ctx, entity); err != nil {
		return nil, err
	}
	return svc.ResolveByID(ctx, cardID)
}

func (svc *Service) UpdateMembers(ctx context.Context, cardID string, members []MemberInput) (*Card, error) {
	entity, err := svc.repo.ResolveByID(ctx, cardID)
	if err != nil {
		return nil, errors.Wrap(err, "resolve card by id")
	}
	err = entity.UpdateMembers(members)
	if err != nil {
		return nil, errors.Wrap(err, "add members")
	}
	err = svc.repo.Store(ctx, entity)
	if err != nil {
		return nil, errors.Wrap(err, "store card")
	}
	return svc.repo.ResolveByID(ctx, cardID)
}

func (svc *Service) AddLabel(ctx context.Context, cardID, labelID string) (*Card, error) {
	cardEntity, err := svc.repo.ResolveByID(ctx, cardID)
	if err != nil {
		return nil, errors.Wrap(err, "resolve card by id")
	}
	exist, err := svc.boardService.ExistLabelByID(ctx, labelID)
	if err != nil {
		return nil, errors.Wrap(err, "exist label by id")
	}
	if !exist {
		return nil, apierror.WithDesc(ErrorCodeEntityNotFound, "label not found")
	}
	err = cardEntity.AddLabel(labelID)
	if err != nil {
		if f, ok := err.(apierror.APIError); ok && f.Code == ErrorCodeAlreadyExist {
			return cardEntity, nil
		}
		return nil, errors.Wrap(err, "add label to card entity")
	}
	err = svc.repo.StoreLabels(ctx, cardEntity.ID, cardEntity.Labels)
	if err != nil {
		return nil, errors.Wrap(err, "store labels")
	}
	return svc.repo.ResolveByID(ctx, cardID)
}

func (svc *Service) RemoveLabel(ctx context.Context, cardID string, labelID string) (*Card, error) {
	cardEntity, err := svc.repo.ResolveByID(ctx, cardID)
	if err != nil {
		return nil, errors.Wrap(err, "resolve card by id")
	}
	cardEntity.RemoveLabel(labelID)
	err = svc.repo.StoreLabels(ctx, cardEntity.ID, cardEntity.Labels)
	if err != nil {
		return nil, errors.Wrap(err, "store labels")
	}
	return svc.repo.ResolveByID(ctx, cardID)
}

func (svc *Service) generateCode(ctx context.Context, retried int) (string, error) {
	letters := []rune("1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	codeLength := 5
	b := make([]rune, codeLength)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	code := string(b)
	total, err := svc.repo.CountByFilter(ctx, Filter{PublicIDs: []string{code}})
	if err != nil {
		return "", errors.Wrap(err, "count rows by public_id")
	}
	if total > 0 && retried <= 3 {
		return svc.generateCode(ctx, retried+1)
	}
	return code, nil
}

func (svc *Service) ResolveByID(ctx context.Context, id string) (*Card, error) {
	return svc.repo.ResolveByID(ctx, id)
}

func (svc *Service) ResolveAllByFilter(ctx context.Context, filter Filter) ([]Card, error) {
	return svc.repo.ResolveAllByFilter(ctx, filter)
}
