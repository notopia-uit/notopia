package grpc

import (
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toTrashed(trash *app.Trashed) *pb.Trashed {
	return &pb.Trashed{
		By: toTrashedBy(trash.By),
		At: timestamppb.New(trash.At),
	}
}

func toTrashedBy(trashedBy app.TrashedBy) pb.TrashedBy {
	switch trashedBy {
	case app.TrashedByPurpose:
		return pb.TrashedBy_TRASHED_BY_PURPOSE
	case app.TrashedByParent:
		return pb.TrashedBy_TRASHED_BY_PARENT
	case app.TrashedByUnspecified:
		return pb.TrashedBy_TRASHED_BY_UNSPECIFIED
	}
	return pb.TrashedBy_TRASHED_BY_UNSPECIFIED
}
