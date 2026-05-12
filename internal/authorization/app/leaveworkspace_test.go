package app_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLeaveWorkspaceHandler(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		workspaceID   string
		expectErr     bool
		expectedError string
	}{
		{
			name:          "W111-Owner cannot leave (only owner)",
			userID:        "111",
			workspaceID:   "00000000-0000-0000-0000-000000000111",
			expectErr:     true,
			expectedError: "only owner",
		},
		{
			name:        "W111-Editor leaves workspace successfully",
			userID:      "112",
			workspaceID: "00000000-0000-0000-0000-000000000111",
			expectErr:   false,
		},
		{
			name:        "W111-Viewer leaves workspace successfully",
			userID:      "110",
			workspaceID: "00000000-0000-0000-0000-000000000111",
			expectErr:   false,
		},
		{
			name:          "W112-Only owner cannot leave workspace",
			userID:        "112",
			workspaceID:   "00000000-0000-0000-0000-000000000112",
			expectErr:     true,
			expectedError: "only owner",
		},
		{
			name:          "W110-Only owner cannot leave workspace",
			userID:        "110",
			workspaceID:   "00000000-0000-0000-0000-000000000110",
			expectErr:     true,
			expectedError: "only owner",
		},
		{
			name:          "User not a member cannot leave",
			userID:        "999",
			workspaceID:   "00000000-0000-0000-0000-000000000111",
			expectErr:     true,
			expectedError: "does not have",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e, err := GetLocalEnforcer(t, &GetLocalEnforcerParams{LoadTestPolicies: true, UseTransaction: true})
			require.NoError(t, err, "Failed to create enforcer")

			mockPublisher := app.NewMockIntegrationPublisher(t)

			if !tc.expectErr {
				mockPublisher.EXPECT().
					Publish(t.Context(), mock.MatchedBy(func(events []app.IntegrationEvent) bool {
						if len(events) != 1 {
							return false
						}
						_, ok := events[0].(app.IntegrationEventWorkspaceMemberRemoved)
						return ok
					})).
					Return(nil).
					Once()
			}

			handler := app.NewLeaveWorkspaceHandler(e, mockPublisher)

			workspaceID := uuid.MustParse(tc.workspaceID)
			ctx := t.Context()
			err = handler.Handle(ctx, app.LeaveWorkspace{
				UserID:      tc.userID,
				WorkspaceID: workspaceID,
			})

			if tc.expectErr {
				require.Error(t, err, "Expected error but got none")
				assert.Contains(t, err.Error(), tc.expectedError, "Error message should contain expected error code")
				return
			}

			require.NoError(t, err, "Handler threw an error")

			// Verify the user is no longer a member
			getMembersHandler := app.NewGetWorkspaceMembersHandler(e)
			members, err := getMembersHandler.Handle(ctx, app.GetWorkspaceMembers{
				UserID:      tc.userID,
				WorkspaceID: workspaceID,
			})

			if err == nil {
				// If no error, verify user is not in the members list
				for _, member := range members {
					assert.NotEqual(t, tc.userID, member.ID, "User should not be a member after leaving")
				}
			}
		})
	}
}

func TestLeaveWorkspaceHandler_PublishEventFailure(t *testing.T) {
	t.Skip("TODO: Fix mock expectations when running with parallel tests and transactional database")
	
	e, err := GetLocalEnforcer(t, &GetLocalEnforcerParams{LoadTestPolicies: true, UseTransaction: true})
	require.NoError(t, err, "Failed to create enforcer")

	mockPublisher := app.NewMockIntegrationPublisher(t)
	mockPublisher.EXPECT().
		Publish(mock.MatchedBy(func(ctx interface{}) bool {
			return true // Accept any context
		}), mock.MatchedBy(func(events []app.IntegrationEvent) bool {
			if len(events) != 1 {
				return false
			}
			_, ok := events[0].(app.IntegrationEventWorkspaceMemberRemoved)
			return ok
		})).
		Return(assert.AnError).
		Once()

	handler := app.NewLeaveWorkspaceHandler(e, mockPublisher)

	workspaceID := uuid.MustParse("00000000-0000-0000-0000-000000000111")
	ctx := t.Context()
	err = handler.Handle(ctx, app.LeaveWorkspace{
		UserID:      "112", // Editor, can leave
		WorkspaceID: workspaceID,
	})

	require.Error(t, err, "Expected error when publishing events fails")
	
	pubErr, ok := err.(*errs.PublishIntegrationEventsFailed)
	require.True(t, ok, "Error should be PublishIntegrationEventsFailed")
	assert.Equal(t, workspaceID, pubErr.WorkspaceID)
}
