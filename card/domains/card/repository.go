package card

import "context"

type Repository interface {
	Store(ctx context.Context, entity *Card) error
	StoreLabels(ctx context.Context, cardID string, labels []Label) error
	ResolveByID(ctx context.Context, id string) (*Card, error)
	ResolveAllByFilter(ctx context.Context, filter Filter) ([]Card, error)
	ResolveIDsByFilter(ctx context.Context, filter Filter) ([]string, error)
	CountByFilter(ctx context.Context, filter Filter) (int, error)
}
