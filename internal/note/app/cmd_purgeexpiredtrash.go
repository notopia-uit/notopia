package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

const purgeMinTrashAge = 24 * time.Hour

type PurgeExpiredTrash struct{}

type PurgeExpiredTrashHandler struct {
	uow          domain.UnitOfWork
	readModel    TrashWorkspacesReadModel
	billingSvc   BillingSvc
	trashService *domain.TrashService
	identity     *commonconfig.Identity
}

func NewPurgeExpiredTrashHandler(
	uow domain.UnitOfWork,
	readModel TrashWorkspacesReadModel,
	billingSvc BillingSvc,
	trashService *domain.TrashService,
	identity *commonconfig.Identity,
) *PurgeExpiredTrashHandler {
	return &PurgeExpiredTrashHandler{
		uow:          uow,
		readModel:    readModel,
		billingSvc:   billingSvc,
		trashService: trashService,
		identity:     identity,
	}
}

var ProvidePurgeExpiredTrashHandler = NewPurgeExpiredTrashHandler

type PurgeExpiredTrashCmd commonhandler.Cmd[PurgeExpiredTrash]

var _ PurgeExpiredTrashCmd = (*PurgeExpiredTrashHandler)(nil)

type workspaceCutoff struct {
	workspaceID uuid.UUID
	cutoff      time.Time
}

func expiredTrashCutoffs(
	ctx context.Context,
	now time.Time,
	workspaceIDs []uuid.UUID,
	ents map[uuid.UUID]Entitlements,
) []workspaceCutoff {
	cutoffs := make([]workspaceCutoff, 0, len(workspaceIDs))
	for _, workspaceID := range workspaceIDs {
		ent, ok := ents[workspaceID]
		if !ok {
			slog.WarnContext(ctx, "skip trash purge: entitlements missing", slog.String("workspace_id", workspaceID.String()))
			continue
		}
		days := ent.Limit(PlanLimitTrashRetentionDays)
		if days.IsUnlimited() {
			continue
		}
		// Items stay at least purgeMinTrashAge: with retention 0 the cutoff would otherwise be now,
		// purging fresh items depending on whether unrelated old items put the workspace in the batch.
		retention := max(time.Duration(days)*24*time.Hour, purgeMinTrashAge)
		cutoffs = append(cutoffs, workspaceCutoff{
			workspaceID: workspaceID,
			cutoff:      now.Add(-retention),
		})
	}
	return cutoffs
}

// No permission check: cron-only, runs as the note service account.
func (h *PurgeExpiredTrashHandler) Handle(ctx context.Context, _ *PurgeExpiredTrash) error {
	now := time.Now()
	workspaceIDs, err := h.readModel.Handle(ctx, now.Add(-purgeMinTrashAge))
	if err != nil {
		return err
	}
	if len(workspaceIDs) == 0 {
		return nil
	}
	ents, err := h.billingSvc.BatchGetEntitlements(ctx, workspaceIDs)
	if err != nil {
		return err
	}
	for _, c := range expiredTrashCutoffs(ctx, now, workspaceIDs, ents) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := h.purgeWorkspace(ctx, c.workspaceID, c.cutoff); err != nil {
			slog.ErrorContext(ctx, "trash purge failed",
				slog.String("workspace_id", c.workspaceID.String()),
				slog.Any("error", err),
			)
		}
	}
	return nil
}

func (h *PurgeExpiredTrashHandler) purgeWorkspace(ctx context.Context, workspaceID uuid.UUID, cutoff time.Time) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		trashedNotes, err := r.Note().GetMany(ctx, &domain.NoteRepoGetManyParams{
			IDs:         nil,
			WorkspaceID: workspaceID,
			TrashOnly:   true,
			TrashedBy:   domain.TrashedByPurpose,
			ForUpdate:   true,
		})
		if err != nil {
			return err
		}
		trashedFolders, err := r.Folder().GetMany(ctx, &domain.FolderRepoGetManyParams{
			IDs:         nil,
			WorkspaceID: workspaceID,
			TrashOnly:   true,
			TrashedBy:   domain.TrashedByPurpose,
			ForUpdate:   true,
		})
		if err != nil {
			return err
		}

		var noteIDs, folderIDs []uuid.UUID
		for _, n := range trashedNotes {
			if n.TrashedAt().Before(cutoff) {
				noteIDs = append(noteIDs, n.ID())
			}
		}
		for _, f := range trashedFolders {
			if f.TrashedAt().Before(cutoff) {
				folderIDs = append(folderIDs, f.ID())
			}
		}
		if len(noteIDs) == 0 && len(folderIDs) == 0 {
			return nil
		}

		targets, err := loadPermanentDeleteTargets(ctx, r, noteIDs, folderIDs)
		if err != nil {
			return err
		}
		targets.delete(h.trashService, h.identity.ServiceAccountUserID)
		return savePermanentDeleteTargets(ctx, r, targets)
	})
}
