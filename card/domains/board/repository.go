package board

import "context"

type Repository interface {
	Store(ctx context.Context, entity *Board) error
	StoreMember(ctx context.Context, entity BoardMember) error
	StoreList(ctx context.Context, entity BoardList) error
	ResolveByID(ctx context.Context, id string) (*Board, error)
	ResolveAllByFilter(ctx context.Context, filter Filter) ([]Board, error)
	ResolveAll(ctx context.Context, offset, limit int) ([]Board, error)
	ResolveTotal(ctx context.Context) (int, error)
}
