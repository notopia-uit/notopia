package errs

import "github.com/google/uuid"

var (
	CodeWorkspaceEventPubSubFailedToCreateMessage Code = "workspaceeventpubsub_1"
	CodeWorkspaceEventPubSubPublishFailed         Code = "workspaceeventpubsub_2"
	CodeWorkspaceEventPubSubSubscribeFailed       Code = "workspaceeventpubsub_3"
)

type WorkspaceEventPubSubFailedToCreateMessage struct {
	Err
	UserID      string
	WorkspaceID uuid.UUID
}

func NewWorkspaceEventPubSubFailedToCreateMessage(
	userID string,
	workspaceID uuid.UUID,
	err error,
) *WorkspaceEventPubSubFailedToCreateMessage {
	return &WorkspaceEventPubSubFailedToCreateMessage{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Err: Err{
			message: "failed to create message for workspace event",
			code:    CodeWorkspaceEventPubSubFailedToCreateMessage,
			err:     err,
		},
	}
}

type WorkspaceEventPubSubPublishFailed struct {
	Err
	UserID      string
	WorkspaceID uuid.UUID
}

func NewWorkspaceEventPubSubPublishFailed(
	userID string,
	workspaceID uuid.UUID,
	err error,
) *WorkspaceEventPubSubPublishFailed {
	return &WorkspaceEventPubSubPublishFailed{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Err: Err{
			message: "failed to publish workspace event",
			code:    CodeWorkspaceEventPubSubPublishFailed,
			err:     err,
		},
	}
}

type WorkspaceEventPubSubSubscribeFailed struct {
	Err
	UserID      string
	WorkspaceID uuid.UUID
}

func NewWorkspaceEventPubSubSubscribeFailed(
	userID string,
	workspaceID uuid.UUID,
	err error,
) *WorkspaceEventPubSubSubscribeFailed {
	return &WorkspaceEventPubSubSubscribeFailed{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Err: Err{
			message: "failed to subscribe to workspace events",
			code:    CodeWorkspaceEventPubSubSubscribeFailed,
			err:     err,
		},
	}
}
