package logic

import (
	"context"

	"notice/rpc/internal/svc"
	"notice/rpc/notice/notice"

	"github.com/zeromicro/go-zero/core/logx"
)

type NoticeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NoticeLogic {
	return &NoticeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *NoticeLogic) Notice(in *notice.NoticeRequest) (*notice.NoticeResponse, error) {
	// todo: add your logic here and delete this line

	return &notice.NoticeResponse{}, nil
}
