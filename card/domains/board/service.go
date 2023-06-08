package board

import (
	"context"

	"github.com/pkg/errors"
)

type Service struct {
	repo      Repository
	labelRepo LabelRepository
}

func NewService(repo Repository, labelRepo LabelRepository) *Service {
	return &Service{repo: repo, labelRepo: labelRepo}
}

func (svc *Service) Create(ctx context.Context, input Input) (res *Board, err error) {
	entity, err := input.ToEntity()
	if err != nil {
		return
	}
	err = svc.repo.Store(ctx, entity)
	if err != nil {
		err = errors.Wrap(err, "store board")
		return
	}
	return svc.repo.ResolveByID(ctx, entity.ID)
}

func (svc *Service) Update(ctx context.Context, id string, input UpdateInput) (res *Board, err error) {
	entity, err := svc.repo.ResolveByID(ctx, id)
	if err != nil {
		err = errors.Wrap(err, "resolve by id")
		return
	}
	err = entity.Update(input)
	if err != nil {
		err = errors.Wrap(err, "update board")
		return
	}
	err = svc.repo.Store(ctx, entity)
	if err != nil {
		err = errors.Wrap(err, "store board")
		return
	}
	return svc.repo.ResolveByID(ctx, entity.ID)
}

func (svc *Service) AddMember(ctx context.Context, boardID string, input MemberListInput) (res *Board, err error) {
	boardEntity, err := svc.repo.ResolveByID(ctx, boardID)
	if err != nil {
		err = errors.Wrap(err, "resolve board by id")
		return
	}
	// TODO(raka) Fix this
	for _, member := range input.Members {
		exist := boardEntity.MemberExist(member.UserID)
		if exist {
			continue
		}
		boardMember, err := NewBoardMember(boardID, member.UserID)
		if err != nil {
			return res, err
		}
		err = svc.repo.StoreMember(ctx, boardMember)
		if err != nil {
			err = errors.Wrap(err, "store member")
			return res, err
		}
	}
	return svc.repo.ResolveByID(ctx, boardID)
}

func (svc *Service) RemoveMember(ctx context.Context, boardID string, userID string) (res *Board, err error) {
	boardEntity, err := svc.repo.ResolveByID(ctx, boardID)
	if err != nil {
		err = errors.Wrap(err, "resolve board by id")
		return
	}
	boardEntity.RemoveMember(userID)
	err = svc.repo.Store(ctx, boardEntity)
	if err != nil {
		err = errors.Wrap(err, "store board entity")
		return
	}
	return svc.repo.ResolveByID(ctx, boardID)
}

func (svc *Service) AddList(ctx context.Context, boardID string, listInput ListInput) (res *Board, err error) {
	_, err = svc.repo.ResolveByID(ctx, boardID)
	if err != nil {
		err = errors.Wrap(err, "resolve board by id")
		return
	}
	boardList, err := NewBoardList(boardID, listInput)
	if err != nil {
		return
	}
	err = svc.repo.StoreList(ctx, boardList)
	if err != nil {
		return
	}
	return svc.repo.ResolveByID(ctx, boardID)
}

func (svc *Service) CreateLabel(ctx context.Context, boardID string, input LabelInput) (res *Board, err error) {
	boardEntity, err := svc.repo.ResolveByID(ctx, boardID)
	if err != nil {
		err = errors.Wrap(err, "resolve by id")
		return
	}
	labelEntity, err := input.ToEntity(boardEntity.ID)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	err = svc.labelRepo.Store(ctx, labelEntity)
	if err != nil {
		err = errors.Wrap(err, "store label")
		return
	}
	return svc.repo.ResolveByID(ctx, boardID)
}

func (svc *Service) UpdateLabel(ctx context.Context, labelID, title, color string) (res *Board, err error) {
	label, err := svc.labelRepo.ResolveByID(ctx, labelID)
	if err != nil {
		err = errors.Wrap(err, "resolve label by id")
		return
	}
	label.Update(title, color)
	err = svc.labelRepo.Store(ctx, label)
	if err != nil {
		err = errors.Wrap(err, "store label")
		return
	}
	return svc.repo.ResolveByID(ctx, label.BoardID)
}

func (svc *Service) ResolveByID(ctx context.Context, id string) (*Board, error) {
	return svc.repo.ResolveByID(ctx, id)
}

func (svc *Service) ExistLabelByID(ctx context.Context, labelID string) (bool, error) {
	return svc.labelRepo.ExistByID(ctx, labelID)
}

func (svc *Service) ResolveAllByFilter(ctx context.Context, filter Filter) ([]Board, error) {
	if filter.IsEmpty() {
		return nil, nil
	}
	return svc.repo.ResolveAllByFilter(ctx, filter)
}

func (svc *Service) ResolvePage(ctx context.Context, pageNum int, pageSize int) (res Page[Board], err error) {
	total, err := svc.repo.ResolveTotal(ctx)
	if err != nil {
		return
	}
	offset := (pageNum - 1) * pageSize
	items, err := svc.repo.ResolveAll(ctx, offset, pageSize)
	if err != nil {
		return
	}
	return Page[Board]{
		Items: items,
		Total: total,
	}, nil
}

func (svc *Service) generatePublicID(ctx context.Context, retried int) (res string, err error) {
	return
}
