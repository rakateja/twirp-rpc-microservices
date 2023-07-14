package card

import (
	"context"
	"encoding/json"
	"fmt"

	redis "github.com/redis/go-redis/v9"
)

type cachedRepository struct {
	sqlRepo Repository
	client  *redis.Client
}

const (
	cachedKey = "card:%s"
)

func NewCachedRepository(repo Repository, redisClient *redis.Client) Repository {
	return &cachedRepository{repo, redisClient}
}

func (repo *cachedRepository) Store(ctx context.Context, entity *Card) error {
	err := repo.sqlRepo.Store(ctx, entity)
	if err != nil {
		return err
	}
	err = repo.client.Del(ctx, fmt.Sprintf(cachedKey, entity.ID)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (repo *cachedRepository) StoreLabels(ctx context.Context, cardID string, labels []Label) error {
	err := repo.sqlRepo.StoreLabels(ctx, cardID, labels)
	if err != nil {
		return err
	}
	err = repo.client.Del(ctx, fmt.Sprintf(cachedKey, cardID)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (repo *cachedRepository) ResolveByID(ctx context.Context, id string) (*Card, error) {
	val, err := repo.client.Get(ctx, fmt.Sprintf(cachedKey, id)).Result()
	if err != nil {
		if err != redis.Nil {
			return nil, err
		}
	} else {
		var res Card
		if err := json.Unmarshal([]byte(val), &res); err != nil {
			return nil, err
		}
		return &res, nil
	}
	res, err := repo.sqlRepo.ResolveByID(ctx, id)
	if err != nil {
		return nil, err
	}
	bt, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	err = repo.client.Set(ctx, fmt.Sprintf(cachedKey, id), string(bt), 0).Err()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (repo *cachedRepository) ResolveAllIDsByFilter(ctx context.Context, filter Filter) ([]string, error) {
	return repo.sqlRepo.ResolveAllIDsByFilter(ctx, filter)
}

func (repo *cachedRepository) ResolveAllByFilter(ctx context.Context, filter Filter) (res []Card, err error) {
	ids, err := repo.sqlRepo.ResolveAllIDsByFilter(ctx, filter)
	if err != nil {
		return
	}
	var keys []string
	for _, id := range ids {
		keys = append(keys, fmt.Sprintf(cachedKey, id))
	}
	if len(keys) == 0 {
		return
	}
	bts, err := repo.client.MGet(ctx, keys...).Result()
	if err != nil {
		return
	}
	cardMap := make(map[string]Card, 0)
	for _, bt := range bts {
		if bt == nil {
			continue
		}
		var card Card
		if err := json.Unmarshal([]byte(bt.(string)), &card); err != nil {
			return res, err
		}
		cardMap[card.ID] = card
	}
	var noCached []string
	for _, id := range ids {
		_, exist := cardMap[id]
		if !exist {
			noCached = append(noCached, id)
		}
	}
	cards, err := repo.sqlRepo.ResolveAllByFilter(ctx, Filter{IDs: noCached})
	if err != nil {
		return
	}
	for _, val := range cardMap {
		cards = append(cards, val)
	}
	return cards, nil
}

func (repo *cachedRepository) ResolveIDsByFilter(ctx context.Context, filter Filter, limit int) ([]string, error) {
	return repo.sqlRepo.ResolveIDsByFilter(ctx, filter, limit)
}

func (repo *cachedRepository) CountByFilter(ctx context.Context, filter Filter) (int, error) {
	return repo.sqlRepo.CountByFilter(ctx, filter)
}
