package domain

import (
	"time"

	"github.com/google/uuid"
)

type BillingAccount struct {
	eventRecorder
	workspaceID        uuid.UUID
	providerCustomerID string
	payer              Payer
	payerMissingSince  *time.Time
	createdAt          time.Time
	updatedAt          time.Time
}

func OpenBillingAccount(workspaceID uuid.UUID, providerCustomerID string, payer Payer, now time.Time, userID string) *BillingAccount {
	a := &BillingAccount{
		eventRecorder:      eventRecorder{events: nil},
		workspaceID:        workspaceID,
		providerCustomerID: providerCustomerID,
		payer:              payer,
		payerMissingSince:  nil,
		createdAt:          now,
		updatedAt:          now,
	}
	a.record(&BillingAccountOpenedEvent{BaseEvent: NewBaseEvent(workspaceID.String(), userID)})
	return a
}

func UnmarshalBillingAccount(workspaceID uuid.UUID, providerCustomerID string, payer Payer, payerMissingSince *time.Time, createdAt, updatedAt time.Time) *BillingAccount {
	return &BillingAccount{
		eventRecorder:      eventRecorder{events: nil},
		workspaceID:        workspaceID,
		providerCustomerID: providerCustomerID,
		payer:              payer,
		payerMissingSince:  payerMissingSince,
		createdAt:          createdAt,
		updatedAt:          updatedAt,
	}
}

func (a *BillingAccount) ChangePayer(payer Payer, now time.Time, userID string) {
	a.payer = payer
	a.payerMissingSince = nil
	a.updatedAt = now
	a.record(&PayerChangedEvent{BaseEvent: NewBaseEvent(a.workspaceID.String(), userID), PayerID: payer.ID})
}

func (a *BillingAccount) MarkPayerMissing(now time.Time, userID string) {
	if a.payerMissingSince != nil {
		return
	}
	a.payerMissingSince = &now
	a.updatedAt = now
	a.record(&PayerMissingEvent{BaseEvent: NewBaseEvent(a.workspaceID.String(), userID)})
}

func (a *BillingAccount) UpdateContact(email, name string, now time.Time, userID string) {
	a.payer.Email, a.payer.Name = email, name
	a.updatedAt = now
	a.record(&BillingContactUpdatedEvent{BaseEvent: NewBaseEvent(a.workspaceID.String(), userID)})
}

func (a *BillingAccount) WorkspaceID() uuid.UUID        { return a.workspaceID }
func (a *BillingAccount) ProviderCustomerID() string    { return a.providerCustomerID }
func (a *BillingAccount) Payer() Payer                  { return a.payer }
func (a *BillingAccount) PayerMissingSince() *time.Time { return a.payerMissingSince }
func (a *BillingAccount) CreatedAt() time.Time          { return a.createdAt }
func (a *BillingAccount) UpdatedAt() time.Time          { return a.updatedAt }
