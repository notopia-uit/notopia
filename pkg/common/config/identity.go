package commonconfig

type Identity struct {
	AdminGroup           string `default:"notopia-admins" json:"adminGroup"           mapstructure:"admin_group"             validate:"required" yaml:"admin_group"`
	ServiceAccountUserID string `default:"-"              json:"serviceAccountUserId" mapstructure:"service_account_user_id" validate:"required" yaml:"service_account_user_id"`
}
