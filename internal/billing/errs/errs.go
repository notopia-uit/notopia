package errs

import "fmt"

type Code string

func (c Code) String() string { return string(c) }

const (
	CodeUnauthorized              Code = "unauthorized"
	CodeForbidden                 Code = "forbidden"
	CodeInvalid                   Code = "invalid"
	CodeNotFound                  Code = "notFound"
	CodeInternal                  Code = "internal"
	CodeSubscriptionAlreadyActive Code = "subscriptionAlreadyActive"
	CodeNoBillingAccount          Code = "noBillingAccount"
	CodeNoActiveSubscription      Code = "noActiveSubscription"
	CodeRefundWindowExpired       Code = "refundWindowExpired"
	CodeAlreadyPayer              Code = "alreadyPayer"
	CodeNotOwner                  Code = "notOwner"
	CodePriceNotActive            Code = "priceNotActive"
	CodePaymentProvider           Code = "paymentProviderError"
	CodeAuthorizationService      Code = "authorizationServiceError"
	CodeWebhookSignature          Code = "webhookSignatureInvalid"
)

type Error interface {
	error
	Code() Code
	Message() string
	Unwrap() error
}

type Err struct {
	code    Code
	message string
	err     error
}

var _ Error = (*Err)(nil)

func New(code Code, message string, err error) *Err {
	return &Err{code: code, message: message, err: err}
}

func (e *Err) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.message, e.err)
	}
	return e.message
}

func (e *Err) Unwrap() error   { return e.err }
func (e *Err) Code() Code      { return e.code }
func (e *Err) Message() string { return e.message }

func NewInvalid(msg string) *Err   { return New(CodeInvalid, msg, nil) }
func NewForbidden(msg string) *Err { return New(CodeForbidden, msg, nil) }
func NewNotFound(msg string) *Err  { return New(CodeNotFound, msg, nil) }
func NewInternal(msg string, err error) *Err {
	return New(CodeInternal, msg, err)
}
