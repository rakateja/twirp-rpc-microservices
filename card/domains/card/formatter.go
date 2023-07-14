package card

import (
	"time"

	timestampPb "github.com/golang/protobuf/ptypes/timestamp"
	pb "github.com/rakateja/milo/twirp-rpc-examples/card/proto/rpcproto"
)

func ToCardPagePb(t CardPage) *pb.CardPage {
	return &pb.CardPage{
		Items: ToCardListPb(t.Items),
		Total: t.Total,
	}
}

func ToCardPb(t Card) *pb.Card {
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

func ToCardListPb(ls []Card) []*pb.Card {
	var res []*pb.Card
	for _, t := range ls {
		res = append(res, ToCardPb(t))
	}
	return res
}

func ToCardMembers(ls []Member) (res []*pb.CardMember) {
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

func ToCardAttachments(ls []Attachment) (res []*pb.CardAttachment) {
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

func ToCardInput(pbInput *pb.CardInput) CardInput {
	var dueDateFrom, dueDateUntil *string
	if pbInput.DueDateFrom == "" {
		dueDateFrom = &pbInput.DueDateFrom
	}
	if pbInput.DueDateUntil == "" {
		dueDateUntil = &pbInput.DueDateUntil
	}
	return CardInput{
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

func ToCardMemberInputFromPb(ls []*pb.AddMemberInput) []MemberInput {
	res := make([]MemberInput, 0)
	for _, inputPb := range ls {
		res = append(res, MemberInput{
			UserID: inputPb.UserId,
		})
	}
	return res
}

func ToTimestampPb(ts *time.Time) *timestampPb.Timestamp {
	if ts == nil {
		return nil
	}
	return &timestampPb.Timestamp{
		Seconds: ts.Unix(),
		Nanos:   int32(ts.Nanosecond()),
	}
}
