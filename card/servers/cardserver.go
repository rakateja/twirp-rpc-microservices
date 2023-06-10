package servers

import (
	"context"

	"github.com/rakateja/milo/twirp-rpc-examples/card/domains/card"
	pb "github.com/rakateja/milo/twirp-rpc-examples/card/proto/rpcproto"
)

type CardServer struct {
	cardSvc *card.Service
}

func NewCardServer(svc *card.Service) pb.CardService {
	return &CardServer{cardSvc: svc}
}

func (svc *CardServer) Create(ctx context.Context, input *pb.CardInput) (*pb.Card, error) {
	res, err := svc.cardSvc.Create(ctx, ToCardInput(input))
	if err != nil {
		return nil, err
	}
	return ToCardPb(*res), nil
}

func (svc *CardServer) Update(ctx context.Context, updateInput *pb.CardUpdateInput) (*pb.Card, error) {
	res, err := svc.cardSvc.Update(ctx, updateInput.Id, ToCardInput(updateInput.Input))
	if err != nil {
		return nil, err
	}
	return ToCardPb(*res), nil
}

func (svc *CardServer) MoveList(ctx context.Context, input *pb.CardMoveListInput) (*pb.Card, error) {
	res, err := svc.cardSvc.MoveList(ctx, input.CardID, input.ListID)
	if err != nil {
		return nil, err
	}
	return ToCardPb(*res), nil
}

func (svc *CardServer) GetByID(ctx context.Context, input *pb.GetByIDInput) (*pb.Card, error) {
	res, err := svc.cardSvc.ResolveByID(ctx, input.Id)
	if err != nil {
		return nil, err
	}
	return ToCardPb(*res), nil
}

func (svc *CardServer) GetPage(ctx context.Context, input *pb.GetPageInput) (*pb.CardPage, error) {
	return nil, nil
}

func (svc *CardServer) GetAll(ctx context.Context, filter *pb.CardFilter) (*pb.CardList, error) {
	res, err := svc.cardSvc.ResolveAllByFilter(ctx, card.Filter{IDs: filter.Ids})
	if err != nil {
		return nil, err
	}
	return &pb.CardList{Cards: ToCardListPb(res)}, nil
}

func ToCardPb(t card.Card) *pb.Card {
	return &pb.Card{
		Id:                 t.ID,
		ListId:             t.ListID,
		PublicId:           t.PublicID,
		Title:              t.Title,
		Description:        t.Description,
		DueDateFrom:        ToTimestampPb(t.DueDateFrom),
		DueDateUntil:       ToTimestampPb(t.DueDateUntil),
		DueDateCompletedAt: ToTimestampPb(t.DueDateCompletedAt),
		Members:            ToCardMembers(t.Members),
		Attachments:        ToCardAttachments(t.Attachments),
		CreatedAt:          ToTimestampPb(&t.CreatedAt),
		UpdatedAt:          ToTimestampPb(&t.UpdatedAt),
		DeletedAt:          ToTimestampPb(t.DeletedAt),
	}
}

func ToCardListPb(ls []card.Card) []*pb.Card {
	var res []*pb.Card
	for _, t := range ls {
		res = append(res, ToCardPb(t))
	}
	return res
}

func ToCardMembers(ls []card.Member) (res []*pb.CardMember) {
	for _, t := range ls {
		res = append(res, &pb.CardMember{
			Id:        t.ID,
			CardId:    t.CardID,
			UserId:    t.UserID,
			CreatedAt: ToTimestampPb(&t.CreatedAt),
		})
	}
	return
}

func ToCardAttachments(ls []card.Attachment) (res []*pb.CardAttachment) {
	for _, t := range ls {
		res = append(res, &pb.CardAttachment{
			Id:        t.ID,
			CardId:    t.CardID,
			LinkName:  t.LinkName,
			FileType:  t.FileType,
			FileUrl:   t.FileURL,
			CreatedAt: ToTimestampPb(&t.CreatedAt),
			UpdatedAt: ToTimestampPb(&t.UpdatedAt),
		})
	}
	return
}

func ToCardInput(pbInput *pb.CardInput) card.CardInput {
	var dueDateFrom, dueDateUntil *string
	if pbInput.DueDateFrom == "" {
		dueDateFrom = &pbInput.DueDateFrom
	}
	if pbInput.DueDateUntil == "" {
		dueDateUntil = &pbInput.DueDateUntil
	}
	return card.CardInput{
		ListID:             pbInput.ListId,
		BoardID:            pbInput.BoardId,
		Title:              pbInput.Title,
		Description:        pbInput.Description,
		DueDateFrom:        dueDateFrom,
		DueDateUntil:       dueDateUntil,
		DueDateIsCompleted: pbInput.DueDateIsCompleted,
		Members:            ToCardMemberInputFromPb(pbInput.Members),
	}
}

func ToCardMemberInputFromPb(ls []*pb.AddMemberInput) []card.MemberInput {
	res := make([]card.MemberInput, 0)
	for _, inputPb := range ls {
		res = append(res, card.MemberInput{
			UserID: inputPb.UserId,
		})
	}
	return res
}
