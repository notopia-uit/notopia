package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/pb"
)

func parseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, errs.NewInvalid("invalid uuid: " + s)
	}
	return id, nil
}

func (s *ServiceServer) GetWorkspaceEntitlements(ctx context.Context, req *pb.GetWorkspaceEntitlementsRequest) (*pb.GetWorkspaceEntitlementsResponse, error) {
	id, err := parseUUID(req.WorkspaceId)
	if err != nil {
		return nil, err
	}
	ent, err := s.app.Queries.GetEntitlements.Handle(ctx, &app.GetEntitlements{WorkspaceID: id, UserID: "", Internal: true})
	if err != nil {
		return nil, err
	}
	out, err := toPBEntitlements(ent)
	if err != nil {
		return nil, err
	}
	return &pb.GetWorkspaceEntitlementsResponse{Entitlements: out}, nil
}

func (s *ServiceServer) BatchGetWorkspaceEntitlements(ctx context.Context, req *pb.BatchGetWorkspaceEntitlementsRequest) (*pb.BatchGetWorkspaceEntitlementsResponse, error) {
	ids := make([]uuid.UUID, 0, len(req.WorkspaceIds))
	for _, raw := range req.WorkspaceIds {
		id, err := parseUUID(raw)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	res, err := s.app.Queries.BatchGetEntitlements.Handle(ctx, &app.BatchGetEntitlements{WorkspaceIDs: ids})
	if err != nil {
		return nil, err
	}
	out := make(map[string]*pb.Entitlements, len(res))
	for id, e := range res {
		if out[id.String()], err = toPBEntitlements(e); err != nil {
			return nil, err
		}
	}
	return &pb.BatchGetWorkspaceEntitlementsResponse{Entitlements: out}, nil
}

func (s *ServiceServer) GetWorkspacePayer(ctx context.Context, req *pb.GetWorkspacePayerRequest) (*pb.GetWorkspacePayerResponse, error) {
	id, err := parseUUID(req.WorkspaceId)
	if err != nil {
		return nil, err
	}
	v, err := s.app.Queries.GetWorkspacePayer.Handle(ctx, &app.GetWorkspacePayer{WorkspaceID: id})
	if err != nil {
		return nil, err
	}
	return &pb.GetWorkspacePayerResponse{PayerId: v.PayerID, RequiresPayer: v.RequiresPayer}, nil
}

func (s *ServiceServer) ListPlans(ctx context.Context, _ *pb.ListPlansRequest) (*pb.ListPlansResponse, error) {
	views, err := s.app.Queries.ListPlans.Handle(ctx, &app.ListPlans{})
	if err != nil {
		return nil, err
	}
	out := &pb.ListPlansResponse{}
	for _, v := range views {
		tier, err := toPBPlanTier(v.Plan.Tier)
		if err != nil {
			return nil, err
		}
		out.Plans = append(out.Plans, &pb.BillingPlan{Tier: tier, Features: v.Plan.Features, Limits: v.Plan.Limits})
	}
	return out, nil
}
