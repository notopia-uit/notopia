---
url: /notopia/pr/428/services/billing/domain.md
---

# Billing Class Diagram

:::info

* Golang syntax
* Apply DDD, CQRS, repository pattern, clean architecture
* Derived from `internal/billing/domain` and `internal/billing/app`

:::

## Domain Layer

`BillingAccount` owns the payer's recorded contact details; `Subscription` owns its
refunds and references a catalog `Price`. The subscription's plan and interval are the
purchased terms. Provider snapshots describe Stripe state, and payment coverage belongs
to each payment, not the subscription's current period. A refund keeps its payment
reference and amount after the temporary quote is gone.

```plantuml
@startuml Billing
title Billing Domain

!$rosewater = "#dc8a78"
!$flamingo  = "#dd7878"
!$pink      = "#ea76cb"
!$mauve     = "#8839ef"
!$red       = "#d20f39"
!$maroon    = "#e64553"
!$peach     = "#fe640b"
!$yellow    = "#df8e1d"
!$green     = "#40a02b"
!$teal      = "#179299"
!$sky       = "#04a5e5"
!$sapphire  = "#209fb5"
!$lavender  = "#7287fd"
!$blue      = "#1e66f5"
!$text      = "#4c4f69"
!$subtext1  = "#5c5f77"
!$subtext0  = "#6c6f85"
!$overlay2  = "#7c7f93"
!$overlay1  = "#8c8fa1"
!$overlay0  = "#9ca0b0"
!$surface2  = "#acb0be"
!$surface1  = "#bcc0cc"
!$surface0  = "#ccd0da"
!$base      = "#eff1f5"
!$mantle    = "#e6e9ef"
!$crust     = "#dce0e8"

skinparam backgroundColor $base
skinparam defaultFontColor $text
skinparam roundcorner 16
skinparam classFontStyle bold

skinparam ArrowColor $subtext0
skinparam packageStyle rectangle

skinparam package {
    BackgroundColor $mantle
    BorderColor $surface1
    FontColor $mauve
}

skinparam class {
    BackgroundColor $base
    BorderColor $surface1
    HeaderBackgroundColor $surface0
    AttributeFontColor $text
}

!define RepoInterface(name) interface "name" as name <<(I, $pink) Repo>>
!define ServiceInterface(name) interface "name" as name <<(I, $rosewater) Service>>
!define Enum(name) enum "name" as name <<(E, $flamingo) Enum>>
!define ValueObject(name) class "name" as name <<(S, $yellow) Value Object>>
!define AggregateRoot(name) class "name" as name <<(S, $sky) Aggregate Root>>
!define RepoParam(name) class "name" as name <<(S, $green) Repo Param>>

package "Enums" {
Enum(PlanTier) {
    UNSPECIFIED
    FREE
    PLUS
    PRO
}

Enum(Interval) {
    UNSPECIFIED
    MONTH
    YEAR
    LIFETIME
}

Enum(SubscriptionStatus) {
    NONE
    INCOMPLETE
    ACTIVE
    PAST_DUE
    UNPAID
    CANCELED
    INCOMPLETE_EXPIRED
}

Enum(RefundReason) {
    UNSPECIFIED
    CANCEL_NOW
    LIFETIME_WINDOW
    REPLACED_BY_LIFETIME
    WORKSPACE_DELETED
    DUPLICATE_CHECKOUT
    EXTERNAL
}

Enum(RefundStatus) {
    UNSPECIFIED
    PENDING
    SUCCEEDED
    FAILED
    CANCELED
}

Enum(Currency) {
    UNSPECIFIED
    VND
}

Enum(FeatureKey) {
    UNSPECIFIED
    AUTO_COMMIT
    WORKSPACE_GRAPH
}

Enum(LimitKey) {
    UNSPECIFIED
    MEMBERS
    PUBLISHED_NOTES
    NOTE_GRAPH_MAX_DEPTH
    AUTO_COMMIT_MIN_INTERVAL
    REVISION_HISTORY_DAYS
    TRASH_RETENTION_DAYS
    AI_CREDITS_MONTHLY
}
}

ValueObject(Money) {
    amount int64
    currency Currency

    Amount()
    Currency()
    Add()
}

ValueObject(Payer) {
    id string
    email string
    name string

    ID()
    Email()
    Name()
}

ValueObject(ProviderSnapshot) {
    providerSubscriptionID string
    providerPriceID string
    workspaceID uuid.UUID
    status SubscriptionStatus
    currentPeriodStart time.Time
    currentPeriodEnd time.Time
    cancelAtPeriodEnd bool
    canceledAt *time.Time
}

ValueObject(PaymentInfo) {
    amount Money
    paymentRef string
    paidAt time.Time
    coverageStart time.Time
    coverageEnd time.Time
}

ValueObject(RefundLine) {
    paymentRef string
    paidAmount Money
    amount Money
    coverageSeconds int64
    unusedSeconds int64
}

ValueObject(RefundQuote) {
    unusedSeconds int64
    periodSeconds int64
    lines []RefundLine

    Amount()
    Lines()
}

ValueObject(Entitlements) {
    workspaceID uuid.UUID
    plan PlanTier
    status SubscriptionStatus
    accessUntil *time.Time
    features map[FeatureKey]bool
    limits map[LimitKey]int64
}

class "Refund" as Refund <<(E, $teal) Entity>> {
    id uuid.UUID
    providerRefundID *string
    paymentRef string
    amount Money
    reason RefundReason
    status RefundStatus
    failureMessage *string
    createdAt time.Time
    updatedAt time.Time

    AwaitingSubmission()
}

AggregateRoot(Subscription) {
    id uuid.UUID
    workspaceID uuid.UUID
    plan PlanTier
    interval Interval
    priceID uuid.UUID
    status SubscriptionStatus
    providerSubscriptionID *string
    providerPaymentID *string
    currentPeriodStart *time.Time
    currentPeriodEnd *time.Time
    cancelAtPeriodEnd bool
    canceledAt *time.Time
    pastDueSince *time.Time
    paidAt *time.Time
    providerSyncedAt time.Time
    refunds []*Refund

    NewSubscription()
    NewLifetimeSubscription()
    NewDuplicateSubscription()
    NewDuplicateLifetimeSubscription()
    ApplySnapshot()
    ScheduleCancel()
    ResumeCancel()
    QuoteCancelNow()
    CancelNow()
    UpdateRefund()
    FailRefund()
    RecordExternalRefund()
    AwaitingProviderCancel()
    MarkProviderSynced()
    IsCurrent()
    GrantsAccess()
    AccessUntil()
}

AggregateRoot(BillingAccount) {
    workspaceID uuid.UUID
    providerCustomerID string
    payer Payer
    payerMissingSince *time.Time

    OpenBillingAccount()
    ChangePayer()
    MarkPayerMissing()
}

AggregateRoot(Price) {
    id uuid.UUID
    plan PlanTier
    interval Interval
    amount Money
    providerPriceID string
    providerLookupKey string
    active bool
    archivedAt *time.Time

    NewPrice()
    Archive()
}

AggregateRoot(Plan) {
    tier PlanTier
    features map[FeatureKey]bool
    limits map[LimitKey]int64

    NewPlan()
    Update()
    Feature()
    Limit()
}

AggregateRoot(BillingPolicy) {
    pastDueGraceDays int
    lifetimeRefundWindowDays int
    portalConfigurationID *string

    Update()
    SetPortalConfiguration()
}

class "ResolveEntitlements" as ResolveEntitlements <<(F, $lavender) Function>> {
    (sub, plans, policy, now) Entitlements
}

RepoInterface(SubscriptionRepo) {
    GetByID()
    GetCurrentByWorkspace()
    GetByProviderSubscriptionID()
    GetByProviderPaymentID()
    GetByRefundID()
    ListCurrentByWorkspaces()
    ListNeedingReconcile()
    ListCanceledAwaitingProviderCancel()
    ListIDsWithUnsubmittedRefunds()
    Save()
}

RepoInterface(BillingAccountRepo) {
    GetByWorkspace()
    ListByPayer()
    Save()
}

RepoInterface(PriceRepo) {
    GetByID()
    GetByProviderPriceID()
    GetActive()
    ListActive()
    List()
    Save()
}

RepoInterface(PlanRepo) {
    Get()
    ListMap()
    Save()
}

RepoInterface(BillingPolicyRepo) {
    Get()
    Save()
}

RepoInterface(ProviderEventLog) {
    MarkProcessed()
    PruneBefore()
}

Subscription *-- Refund
Subscription ..> ProviderSnapshot : applies
Subscription ..> RefundQuote : quotes
Subscription ..> PaymentInfo : quotes from
Subscription --> Price : priceID
Subscription ..> BillingPolicy : GrantsAccess
RefundQuote *-- RefundLine
RefundLine ..> Refund : creates on cancellation
RefundLine o-- Money
PaymentInfo o-- Money
Refund o-- Money
Price o-- Money
BillingAccount *-- Payer
BillingAccount "1" -- "0..*" Subscription : workspaceID
Price --> Plan : tier
ResolveEntitlements ..> Subscription
ResolveEntitlements ..> Plan
ResolveEntitlements ..> BillingPolicy
ResolveEntitlements ..> Entitlements

SubscriptionRepo -- Subscription
BillingAccountRepo -- BillingAccount
PriceRepo -- Price
PlanRepo -- Plan
BillingPolicyRepo -- BillingPolicy
@enduml
```

## Application Layer

`CancelNow` only records the local cancellation and refund requests. The provider side effects are separate commands driven by billing's own outbox events (see [Flows](./flows#cancel-now)).

```plantuml
@startuml Billing
title Billing App Layer

!$rosewater = "#dc8a78"
!$flamingo  = "#dd7878"
!$pink      = "#ea76cb"
!$mauve     = "#8839ef"
!$red       = "#d20f39"
!$maroon    = "#e64553"
!$peach     = "#fe640b"
!$yellow    = "#df8e1d"
!$green     = "#40a02b"
!$teal      = "#179299"
!$sky       = "#04a5e5"
!$sapphire  = "#209fb5"
!$lavender  = "#7287fd"
!$blue      = "#1e66f5"
!$text      = "#4c4f69"
!$subtext1  = "#5c5f77"
!$subtext0  = "#6c6f85"
!$overlay2  = "#7c7f93"
!$overlay1  = "#8c8fa1"
!$overlay0  = "#9ca0b0"
!$surface2  = "#acb0be"
!$surface1  = "#bcc0cc"
!$surface0  = "#ccd0da"
!$base      = "#eff1f5"
!$mantle    = "#e6e9ef"
!$crust     = "#dce0e8"

skinparam backgroundColor $base
skinparam defaultFontColor $text
skinparam roundcorner 16
skinparam classFontStyle bold

skinparam ArrowColor $subtext0
skinparam packageStyle rectangle

skinparam package {
    BackgroundColor $mantle
    BorderColor $surface1
    FontColor $mauve
}

skinparam class {
    BackgroundColor $base
    BorderColor $surface1
    HeaderBackgroundColor $surface0
    AttributeFontColor $text
}

!define RepoInterface(name) interface "name" as name <<(I, $pink) Repo>>
!define ServiceInterface(name) interface "name" as name <<(I, $rosewater) Service>>
!define Enum(name) enum "name" as name <<(E, $flamingo) Enum>>
!define ValueObject(name) class "name" as name <<(S, $yellow) Value Object>>
!define AggregateRoot(name) class "name" as name <<(S, $sky) Aggregate Root>>
!define RepoParam(name) class "name" as name <<(S, $green) Repo Param>>

class "Cmds" as Cmds <<(S, $blue) Handlers>> {
    ArchivePrice
    CancelNow
    CancelProviderSubscription
    CompleteCheckout
    CreatePrice
    EnsureDefaults
    EnsurePortalConfiguration
    HandlePayerRemoved
    HandleWorkspaceDeleted
    PruneProviderEvents
    ReconcileSubscriptions
    SubmitRefund
    SyncRefund
    SyncSubscription
    UpdatePlan
    UpdatePolicy
}

class "Queries" as Queries <<(S, $blue) Handlers>> {
    BatchGetEntitlements
    ClaimPayer
    GetEntitlements
    GetPlan
    GetPolicy
    GetPrice
    GetWorkspaceBilling
    GetWorkspacePayer
    ListPlans
    ListPrices
    PreviewCancelNow
    StartCheckout
    StartPortal
}

class "SubscriptionCanceller" as SubscriptionCanceller <<(S, $sky) App Service>> {
    Quote()
    CancelNow()
}

ServiceInterface(PaymentProvider) {
    CreateCustomer()
    UpdateCustomerContact()
    CreateCheckoutSession()
    GetCheckoutSession()
    CreatePortalSession()
    GetSubscriptionSnapshot()
    SetCancelAtPeriodEnd()
    CancelSubscriptionNow()
    PeriodPayments()
    PaymentForIntent()
    CreateRefund()
    GetRefund()
    SetDefaultPaymentMethod()
    DetachPaymentMethods()
    CreatePrice()
    ArchivePrice()
    CreatePortalConfiguration()
}

ServiceInterface(AuthorizationSvc) {
    HasWorkspacePermission()
    GetMemberRole()
}

interface "BillingReadModel" as BillingReadModel <<(I, $teal) Read Model>> {
    CurrentSubscription()
    CurrentSubscriptions()
    Account()
    Plans()
    Policy()
    ActivePrices()
    Prices()
}

interface "domain.SubscriptionRepo" as SubscriptionRepo <<(I, $pink) Repo>>

package "Enums" {
Enum(CheckoutPurpose) {
    UNSPECIFIED
    SUBSCRIBE
    LIFETIME
    PAYER_CHANGE
}

Enum(WorkspacePermission) {
    UNSPECIFIED
    BILLING_READ
    BILLING_MANAGE
}

Enum(WorkspaceRole) {
    NONE
    OWNER
    EDITOR
    VIEWER
}
}

ValueObject(CheckoutInput) {
    purpose CheckoutPurpose
    workspaceID uuid.UUID
    customerID string
    payer Payer
    priceID uuid.UUID
    providerPriceID string
    successURL string
    cancelURL string
}

ValueObject(CheckoutSessionResult) {
    id string
    purpose CheckoutPurpose
    complete bool
    workspaceID uuid.UUID
    payer Payer
    priceID uuid.UUID
    customerID string
    providerSubscriptionID string
    paymentMethodID string
    payment PaymentInfo
}

ValueObject(RefundResult) {
    providerRefundID string
    refundID *uuid.UUID
    paymentRef string
    amount Money
    status RefundStatus
    failureMessage *string
    providerSubscriptionID string
}

SubscriptionCanceller --> SubscriptionRepo
SubscriptionCanceller --> PaymentProvider

Cmds --> PaymentProvider
Cmds --> AuthorizationSvc
Cmds --> SubscriptionCanceller : CancelNow
Queries --> BillingReadModel : reads
Queries --> PaymentProvider
Queries --> AuthorizationSvc
Queries --> SubscriptionCanceller : Quote

PaymentProvider ..> CheckoutInput
PaymentProvider ..> CheckoutSessionResult
PaymentProvider ..> RefundResult
@enduml
```
