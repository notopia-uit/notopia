package component

import (
	"fmt"

	"github.com/notopia-uit/notopia/internal/billing/domain"
)

const (
	TopicSubscriptionChanged = "events.integration.billing.subscription.changed"
)

const (
	TopicWorkspaceDeleted = "events.integration.note.workspace.deleted"
	TopicRoleUpdated      = "events.integration.authorization.user_workspace_role_updated"
	TopicMemberRemoved    = "events.integration.authorization.workspace_member_removed"
)

const DomainEventTopicPrefix = "events.internal.billing."

const (
	TopicPlanUpdated                  = DomainEventTopicPrefix + "plan.updated"
	TopicBillingPolicyUpdated         = DomainEventTopicPrefix + "billing_policy.updated"
	TopicPortalConfigurationChanged   = DomainEventTopicPrefix + "billing_policy.portal_configuration_changed"
	TopicPriceCreated                 = DomainEventTopicPrefix + "price.created"
	TopicPriceArchived                = DomainEventTopicPrefix + "price.archived"
	TopicBillingAccountOpened         = DomainEventTopicPrefix + "billing_account.opened"
	TopicBillingAccountPayerChanged   = DomainEventTopicPrefix + "billing_account.payer_changed"
	TopicBillingAccountPayerMissing   = DomainEventTopicPrefix + "billing_account.payer_missing"
	TopicBillingAccountContactUpdated = DomainEventTopicPrefix + "billing_account.contact_updated"
	TopicSubscriptionCreated          = DomainEventTopicPrefix + "subscription.created"
	TopicSubscriptionActivated        = DomainEventTopicPrefix + "subscription.activated"
	TopicSubscriptionRenewed          = DomainEventTopicPrefix + "subscription.renewed"
	TopicSubscriptionPastDue          = DomainEventTopicPrefix + "subscription.past_due"
	TopicSubscriptionPlanChanged      = DomainEventTopicPrefix + "subscription.plan_changed"
	TopicSubscriptionCancelScheduled  = DomainEventTopicPrefix + "subscription.cancel_scheduled"
	TopicSubscriptionCancelResumed    = DomainEventTopicPrefix + "subscription.cancel_resumed"
	TopicSubscriptionCanceledNow      = DomainEventTopicPrefix + "subscription.canceled_now"
	TopicSubscriptionEnded            = DomainEventTopicPrefix + "subscription.ended"
	TopicSubscriptionRefundRequested  = DomainEventTopicPrefix + "subscription.refund_requested"
	TopicSubscriptionRefundSucceeded  = DomainEventTopicPrefix + "subscription.refund_succeeded"
	TopicSubscriptionRefundFailed     = DomainEventTopicPrefix + "subscription.refund_failed"
)

// DomainEventToTopic maps every domain event to the topic the outbox publishes it on.
// SubscriptionStateChangedEvent is published as the integration event share.SubscriptionChangedEvent.
func DomainEventToTopic(event domain.Event) (string, error) {
	switch event.(type) {
	case *domain.PlanUpdatedEvent:
		return TopicPlanUpdated, nil
	case *domain.BillingPolicyUpdatedEvent:
		return TopicBillingPolicyUpdated, nil
	case *domain.PortalConfigurationChangedEvent:
		return TopicPortalConfigurationChanged, nil
	case *domain.PriceCreatedEvent:
		return TopicPriceCreated, nil
	case *domain.PriceArchivedEvent:
		return TopicPriceArchived, nil
	case *domain.BillingAccountOpenedEvent:
		return TopicBillingAccountOpened, nil
	case *domain.PayerChangedEvent:
		return TopicBillingAccountPayerChanged, nil
	case *domain.PayerMissingEvent:
		return TopicBillingAccountPayerMissing, nil
	case *domain.BillingContactUpdatedEvent:
		return TopicBillingAccountContactUpdated, nil
	case *domain.SubscriptionCreatedEvent:
		return TopicSubscriptionCreated, nil
	case *domain.SubscriptionActivatedEvent:
		return TopicSubscriptionActivated, nil
	case *domain.SubscriptionRenewedEvent:
		return TopicSubscriptionRenewed, nil
	case *domain.SubscriptionPastDueEvent:
		return TopicSubscriptionPastDue, nil
	case *domain.SubscriptionPlanChangedEvent:
		return TopicSubscriptionPlanChanged, nil
	case *domain.SubscriptionCancelScheduledEvent:
		return TopicSubscriptionCancelScheduled, nil
	case *domain.SubscriptionCancelResumedEvent:
		return TopicSubscriptionCancelResumed, nil
	case *domain.SubscriptionCanceledNowEvent:
		return TopicSubscriptionCanceledNow, nil
	case *domain.SubscriptionEndedEvent:
		return TopicSubscriptionEnded, nil
	case *domain.SubscriptionStateChangedEvent:
		return TopicSubscriptionChanged, nil
	case *domain.RefundRequestedEvent:
		return TopicSubscriptionRefundRequested, nil
	case *domain.RefundSucceededEvent:
		return TopicSubscriptionRefundSucceeded, nil
	case *domain.RefundFailedEvent:
		return TopicSubscriptionRefundFailed, nil
	default:
		return "", fmt.Errorf("unknown domain event type: %T", event)
	}
}
