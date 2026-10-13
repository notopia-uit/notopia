package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/api/billing"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
)

func statusOf(code errs.Code) int {
	switch code {
	case errs.CodeUnauthorized:
		return http.StatusUnauthorized
	case errs.CodeForbidden, errs.CodeNotOwner:
		return http.StatusForbidden
	case errs.CodeInvalid, errs.CodeWebhookSignature:
		return http.StatusBadRequest
	case errs.CodeNotFound, errs.CodeNoBillingAccount:
		return http.StatusNotFound
	case errs.CodeSubscriptionAlreadyActive, errs.CodeNoActiveSubscription, errs.CodeRefundWindowExpired,
		errs.CodeAlreadyPayer, errs.CodePriceNotActive:
		return http.StatusConflict
	case errs.CodePaymentProvider:
		return http.StatusBadGateway
	case errs.CodeAuthorizationService:
		return http.StatusServiceUnavailable
	case errs.CodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func writeErr(c *gin.Context, status int, code, message string) {
	c.JSON(status, billing.Error{
		Code:     code,
		Message:  message,
		MoreInfo: nil,
	})
}

func ginMiddlewareErrorHandler(c *gin.Context, message string, statusCode int) {
	code := errs.CodeInternal
	switch statusCode {
	case http.StatusUnauthorized:
		code = errs.CodeUnauthorized
	case http.StatusBadRequest:
		code = errs.CodeInvalid
	case http.StatusNotFound:
		code = errs.CodeNotFound
	}
	writeErr(c, statusCode, code.String(), message)
}

var _ ginmiddleware.ErrorHandler = ginMiddlewareErrorHandler

func strictHandlerRequestErrorHandler(c *gin.Context, err error) {
	writeErr(c, http.StatusBadRequest, errs.CodeInvalid.String(), err.Error())
}

func strictHandlerResponseErrorHandler(c *gin.Context, err error) {
	writeErr(c, http.StatusInternalServerError, errs.CodeInternal.String(), err.Error())
}

func strictHandlerError(c *gin.Context, err error) {
	var e errs.Error
	if errors.As(err, &e) {
		status := statusOf(e.Code())
		message := e.Message()
		if status == http.StatusInternalServerError {
			message = "internal error"
		}
		writeErr(c, status, e.Code().String(), message)
		return
	}
	writeErr(c, http.StatusInternalServerError, errs.CodeInternal.String(), "internal error")
}
