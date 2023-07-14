package card

import (
	"context"
	"log"

	pb "github.com/rakateja/milo/twirp-rpc-examples/card/proto/rpcproto"
)

type CardServer struct {
	cardSvc *Service
}

func NewRPCServer(svc *Service) pb.CardService {
	return &CardServer{cardSvc: svc}
}

func (svc *CardServer) Create(ctx context.Context, input *pb.CardInput) (*pb.Card, error) {
	res, err := svc.cardSvc.Create(ctx, ToCardInput(input))
	if err != nil {
		log.Printf("[ERROR] %+v", err)
		return nil, err
	}
	return ToCardPb(*res), nil
}

func (svc *CardServer) Update(ctx context.Context, updateInput *pb.CardUpdateInput) (*pb.Card, error) {
	res, err := svc.cardSvc.Update(ctx, updateInput.Id, ToCardInput(updateInput.Input))
	if err != nil {
		log.Printf("[ERROR] %+v", err)
		return nil, err
	}
	return ToCardPb(*res), nil
}

func (svc *CardServer) MoveList(ctx context.Context, input *pb.CardMoveListInput) (*pb.Card, error) {
	res, err := svc.cardSvc.MoveList(ctx, input.CardID, input.ListID)
	if err != nil {
		log.Printf("[ERROR] %+v", err)
		return nil, err
	}
	return ToCardPb(*res), nil
}

func (svc *CardServer) GetByID(ctx context.Context, input *pb.GetByIDInput) (*pb.Card, error) {
	res, err := svc.cardSvc.ResolveByID(ctx, input.Id)
	if err != nil {
		log.Printf("[ERROR] %+v", err)
		return nil, err
	}
	return ToCardPb(*res), nil
}

func (svc *CardServer) Search(ctx context.Context, input *pb.GetPageInput) (*pb.CardPage, error) {
	var boardIDs []string
	if input.Filter != nil {
		boardIDs = input.Filter.BoardIds
	}
	res, err := svc.cardSvc.Search(ctx, input.Page, input.Limit, Filter{BoardIDs: boardIDs})
	if err != nil {
		log.Printf("[ERROR] %+v", err)
		return nil, err
	}
	return ToCardPagePb(res), nil
}

func (svc *CardServer) GetAll(ctx context.Context, filter *pb.CardFilter) (*pb.CardList, error) {
	res, err := svc.cardSvc.ResolveAllByFilter(ctx, Filter{IDs: filter.Ids})
	if err != nil {
		log.Printf("[ERROR] %+v", err)
		return nil, err
	}
	return &pb.CardList{Cards: ToCardListPb(res)}, nil
}
