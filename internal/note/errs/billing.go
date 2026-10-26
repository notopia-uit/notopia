package errs

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	CodeFeatureNotInPlan          Code = "featureNotInPlan"
	CodePlanLimitExceeded         Code = "planLimitExceeded"
	CodeBillingPayerMustTransfer  Code = "billingPayerMustTransfer"
	CodeBillingServiceUnavailable Code = "billingServiceUnavailable"
)

type PaymentRequiredKind string

const (
	PaymentRequiredKindFeature PaymentRequiredKind = "feature"
	PaymentRequiredKindLimit   PaymentRequiredKind = "limit"
)

type PaymentRequired struct {
	Err
	Kind         PaymentRequiredKind
	Key          string
	Current      *int64
	Max          *int64
	RequiredPlan string
}

func NewFeatureNotInPlan(key, requiredPlan string) *PaymentRequired {
	return &PaymentRequired{
		Err: Err{
			code:    CodeFeatureNotInPlan,
			message: fmt.Sprintf("feature %s is not included in the workspace plan", key),
		},
		Kind:         PaymentRequiredKindFeature,
		Key:          key,
		Current:      nil,
		Max:          nil,
		RequiredPlan: requiredPlan,
	}
}

func NewPlanLimitExceeded(key string, current, max int64, requiredPlan string) *PaymentRequired {
	return &PaymentRequired{
		Err: Err{
			code:    CodePlanLimitExceeded,
			message: fmt.Sprintf("workspace plan limit %s exceeded", key),
		},
		Kind:         PaymentRequiredKindLimit,
		Key:          key,
		Current:      &current,
		Max:          &max,
		RequiredPlan: requiredPlan,
	}
}

func NewBillingPayerMustTransfer(workspaceID uuid.UUID) *Err {
	return &Err{
		code:    CodeBillingPayerMustTransfer,
		message: fmt.Sprintf("the billing payer of workspace %s cannot leave or lose owner role; another owner must claim billing first", workspaceID),
	}
}

func NewBillingServiceUnavailable(err error) *Err {
	return &Err{
		code:    CodeBillingServiceUnavailable,
		message: "billing service unavailable",
		err:     err,
	}
}
