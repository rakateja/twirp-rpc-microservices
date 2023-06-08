package servers

import (
	"context"

	"github.com/rakateja/milo/twirp-rpc-examples/card/domains/board"
	pb "github.com/rakateja/milo/twirp-rpc-examples/card/proto/rpcproto"
	"github.com/twitchtv/twirp"
)

type BoardServer struct {
	boardSvc *board.Service
}

func NewBoardServer(svc *board.Service) pb.BoardService {
	return &BoardServer{boardSvc: svc}
}

func (svc *BoardServer) CreateBoard(ctx context.Context, input *pb.BoardCreateInput) (*pb.Board, error) {
	res, err := svc.boardSvc.Create(ctx, ToBoardInputFromCreateInputPb(input))
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}
	return ToBoardPb(*res), nil
}

func (svc *BoardServer) UpdateBoard(ctx context.Context, input *pb.BoardUpdateInput) (*pb.Board, error) {
	res, err := svc.boardSvc.Update(ctx, input.Id, board.UpdateInput{Title: input.Title})
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}
	return ToBoardPb(*res), nil
}

func (svc *BoardServer) AddMember(context.Context, *pb.AddMemberInput) (*pb.Board, error) {
	return nil, nil
}

func (svc *BoardServer) AddLabel(context.Context, *pb.AddLabelInput) (*pb.Board, error) {
	return nil, nil
}

func (svc *BoardServer) GetByID(ctx context.Context, input *pb.GetByIDInput) (*pb.Board, error) {
	res, err := svc.boardSvc.ResolveByID(ctx, input.Id)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}
	return ToBoardPb(*res), nil
}

func (svc *BoardServer) GetPage(ctx context.Context, input *pb.GetPageInput) (*pb.BoardPage, error) {
	res, err := svc.boardSvc.ResolvePage(ctx, int(input.Page), int(input.Limit))
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}
	return ToBoardPagePb(res), nil
}
