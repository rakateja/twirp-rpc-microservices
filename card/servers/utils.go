package servers

import (
	"time"

	timestampPb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/rakateja/milo/twirp-rpc-examples/card/domains/board"
	pb "github.com/rakateja/milo/twirp-rpc-examples/card/proto/rpcproto"
)

func ToBoardInputFromCreateInputPb(pbInput *pb.BoardCreateInput) board.Input {
	return board.Input{
		Title:   pbInput.Title,
		Members: ToBoardMemberInputFromPb(pbInput.Members),
		Lists:   ToBoardListInputFromPb(pbInput.Lists),
		Labels:  ToBoardLabelInputFromPb(pbInput.Labels),
	}
}

func ToBoardMemberInputFromPb(ls []*pb.AddMemberInput) []board.MemberInput {
	res := make([]board.MemberInput, 0)
	for _, inputPb := range ls {
		res = append(res, board.MemberInput{
			UserID: inputPb.UserId,
		})
	}
	return res
}

func ToBoardLabelInputFromPb(ls []*pb.AddLabelInput) []board.LabelInput {
	res := make([]board.LabelInput, 0)
	for _, inputPb := range ls {
		res = append(res, board.LabelInput{
			Title: inputPb.Name,
			Color: inputPb.Color,
		})
	}
	return res
}

func ToBoardListInputFromPb(ls []*pb.AddListInput) []board.ListInput {
	res := make([]board.ListInput, 0)
	for _, inputPb := range ls {
		res = append(res, board.ListInput{
			Title:    inputPb.Name,
			Position: int(inputPb.Position),
		})
	}
	return res
}

func ToBoardInputFromUpdateInputPb(pbInput *pb.BoardUpdateInput) board.Input {
	return board.Input{
		Title: pbInput.Title,
	}
}

func ToBoardPagePb(boardPage board.Page[board.Board]) *pb.BoardPage {
	var items []*pb.Board
	for _, entity := range boardPage.Items {
		items = append(items, ToBoardPb(entity))
	}
	return &pb.BoardPage{
		Items: items,
		Total: int32(boardPage.Total),
	}
}

func ToBoardPb(entity board.Board) *pb.Board {
	return &pb.Board{
		Id:        entity.ID,
		Title:     entity.Title,
		Members:   ToBoardMembersPb(entity.Members),
		Lists:     ToBoardListsPb(entity.Lists),
		Labels:    ToBoardLabelsPb(entity.Labels),
		CreatedAt: ToTimestampPb(entity.CreatedAt),
		UpdatedAt: ToTimestampPb(entity.UpdatedAt),
	}
}

func ToTimestampPb(ts time.Time) *timestampPb.Timestamp {
	return &timestampPb.Timestamp{
		Seconds: ts.Unix(),
		Nanos:   int32(ts.Nanosecond()),
	}
}

func ToBoardMembersPb(ls []board.BoardMember) []*pb.BoardMember {
	res := make([]*pb.BoardMember, 0)
	for _, entity := range ls {
		res = append(res, &pb.BoardMember{
			Id:        entity.ID,
			BoardId:   entity.BoardID,
			UserId:    entity.UserID,
			CreatedAt: ToTimestampPb(entity.CreatedAt),
		})
	}
	return res
}

func ToBoardListsPb(ls []board.BoardList) []*pb.BoardList {
	res := make([]*pb.BoardList, 0)
	for _, entity := range ls {
		res = append(res, &pb.BoardList{
			Id:        entity.ID,
			BoardId:   entity.BoardID,
			PublicId:  entity.PublicID,
			Title:     entity.Title,
			Position:  int32(entity.Position),
			CreatedAt: ToTimestampPb(entity.CreatedAt),
			UpdatedAt: ToTimestampPb(entity.UpdatedAt),
		})
	}
	return res
}

func ToBoardLabelsPb(ls []board.Label) []*pb.BoardLabel {
	res := make([]*pb.BoardLabel, 0)
	for _, entity := range ls {
		res = append(res, &pb.BoardLabel{
			Id:        entity.ID,
			BoardId:   entity.BoardID,
			Slug:      entity.Slug,
			Title:     entity.Title,
			Color:     entity.Color,
			CreatedAt: ToTimestampPb(entity.CreatedAt),
			UpdatedAt: ToTimestampPb(entity.CreatedAt),
		})
	}
	return res
}
