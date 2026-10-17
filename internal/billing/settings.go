package billing

import (
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/config"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

// NewAppSettings maps config onto app.Settings here, so config never imports app.
func NewAppSettings(billingCfg *config.Billing, identity *commonconfig.Identity) *app.Settings {
	return &app.Settings{
		WebBaseURL:           billingCfg.WebBaseURL,
		ServiceAccountUserID: identity.ServiceAccountUserID,
	}
}

var ProvideAppSettings = NewAppSettings
