package config

import (
	"fmt"
	"time"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/spf13/viper"
)

type Server struct {
	URL    string                     `json:"url"    mapstructure:"url"    validate:"required,url" yaml:"url"`
	Health commonconfig.ServerAddress `json:"health" mapstructure:"health" validate:"required"     yaml:"health"`
	HTTP   commonconfig.ServerAddress `json:"http"   mapstructure:"http"   validate:"required"     yaml:"http"`
	GRPC   commonconfig.ServerAddress `json:"grpc"   mapstructure:"grpc"   validate:"required"     yaml:"grpc"`
}

type Services struct {
	Authorization commonconfig.Service `json:"authorization" mapstructure:"authorization" validate:"required" yaml:"authorization"`
}

type StripeAPI struct {
	SecretKey commonconfig.Secret `default:"-" json:"-" mapstructure:"secret_key" validate:"required" yaml:"secret_key"`
}

type Stripe struct {
	StripeAPI     `                     mapstructure:",squash"`
	WebhookSecret commonconfig.Secret `default:"-" json:"-" mapstructure:"webhook_secret" validate:"required" yaml:"webhook_secret"`
}

type Billing struct {
	Currency   string `default:"vnd" json:"currency"   mapstructure:"currency"     validate:"required"     yaml:"currency"`
	WebBaseURL string `default:"-"   json:"webBaseUrl" mapstructure:"web_base_url" validate:"required,url" yaml:"web_base_url"`
}

type DomainEvent struct {
	MessageMetadataAggregateIDKey  string `default:"aggregate_id"  json:"messageMetadataAggregateIdKey"  mapstructure:"message_metadata_aggregate_id_key"  validate:"required" yaml:"message_metadata_aggregate_id_key"`
	MessageMetadataUserIDKey       string `default:"user_id"       json:"messageMetadataUserIdKey"       mapstructure:"message_metadata_user_id_key"       validate:"required" yaml:"message_metadata_user_id_key"`
	MessageMetadataPartitionKeyKey string `default:"partition_key" json:"messageMetadataPartitionKeyKey" mapstructure:"message_metadata_partition_key_key" validate:"required" yaml:"message_metadata_partition_key_key"`
}

type Cron struct {
	ReconcileSubscriptions commonconfig.CronJob `json:"reconcileSubscriptions" mapstructure:"reconcile_subscriptions" validate:"required" yaml:"reconcile_subscriptions"`
	PruneProviderEvents    commonconfig.CronJob `json:"pruneProviderEvents"    mapstructure:"prune_provider_events"   validate:"required" yaml:"prune_provider_events"`
}

type ServeConfig struct {
	General  commonconfig.General  `json:"general"  mapstructure:"general"  validate:"omitempty" yaml:"general"`
	Log      commonconfig.Log      `json:"log"      mapstructure:"log"      validate:"omitempty" yaml:"log"`
	Server   Server                `json:"server"   mapstructure:"server"   validate:"required"  yaml:"server"`
	Database commonconfig.SQL      `json:"database" mapstructure:"database" validate:"required"  yaml:"database"`
	Kafka    commonconfig.Kafka    `json:"kafka"    mapstructure:"kafka"    validate:"required"  yaml:"kafka"`
	Redis    commonconfig.Redis    `json:"redis"    mapstructure:"redis"    validate:"required"  yaml:"redis"`
	Identity commonconfig.Identity `json:"identity" mapstructure:"identity" validate:"required"  yaml:"identity"`
	Services Services              `json:"services" mapstructure:"services" validate:"required"  yaml:"services"`
	Stripe   Stripe                `json:"-"        mapstructure:"stripe"   validate:"required"  yaml:"stripe"`
	Billing  Billing               `json:"billing"  mapstructure:"billing"  validate:"required"  yaml:"billing"`
	Cron     Cron                  `json:"cron"     mapstructure:"cron"     validate:"required"  yaml:"cron"`

	DomainEvent DomainEvent `json:"domainEvent" mapstructure:"domain_event" validate:"required" yaml:"domain_event"`
}

var (
	defaultReconcileSubscriptions = commonconfig.CronJob{
		Name:     "billing:reconcile-subscriptions",
		Schedule: "0 * * * *",
		Timeout:  15 * time.Minute,
		Enabled:  true,
	}
	defaultPruneProviderEvents = commonconfig.CronJob{
		Name:     "billing:prune-provider-events",
		Schedule: "0 4 * * *",
		Timeout:  5 * time.Minute,
		Enabled:  true,
	}
)

func setCronDefaults(v *viper.Viper, key string, job commonconfig.CronJob) {
	v.SetDefault(key+".name", job.Name)
	v.SetDefault(key+".schedule", job.Schedule)
	v.SetDefault(key+".timeout", job.Timeout)
	v.SetDefault(key+".enabled", job.Enabled)
}

type Config = ServeConfig

type MigrateConfig struct {
	General  commonconfig.General `json:"general"  mapstructure:"general"  validate:"omitempty" yaml:"general"`
	Log      commonconfig.Log     `json:"log"      mapstructure:"log"      validate:"omitempty" yaml:"log"`
	Database commonconfig.SQL     `json:"database" mapstructure:"database" validate:"required"  yaml:"database"`
}

type SeedConfig struct {
	General  commonconfig.General  `json:"general"  mapstructure:"general"  validate:"omitempty" yaml:"general"`
	Log      commonconfig.Log      `json:"log"      mapstructure:"log"      validate:"omitempty" yaml:"log"`
	Database commonconfig.SQL      `json:"database" mapstructure:"database" validate:"required"  yaml:"database"`
	Identity commonconfig.Identity `json:"identity" mapstructure:"identity" validate:"required"  yaml:"identity"`
	Stripe   StripeAPI             `json:"-"        mapstructure:"stripe"   validate:"required"  yaml:"stripe"`
	Billing  Billing               `json:"billing"  mapstructure:"billing"  validate:"required"  yaml:"billing"`
}

func NewMigrateConfig(validate *validator.Validate, v *viper.Viper) (*MigrateConfig, error) {
	return commonconfig.Load[MigrateConfig](validate, v, "notopia_billing", "billing.notopia.config")
}

func NewSeedConfig(validate *validator.Validate, v *viper.Viper) (*SeedConfig, error) {
	return commonconfig.Load[SeedConfig](validate, v, "notopia_billing", "billing.notopia.config")
}

func New(validate *validator.Validate, v *viper.Viper) (*Config, error) {
	setCronDefaults(v, "cron.reconcile_subscriptions", defaultReconcileSubscriptions)
	setCronDefaults(v, "cron.prune_provider_events", defaultPruneProviderEvents)
	v.SetDefault("server.http.port", 8084)
	v.SetDefault("server.grpc.port", 18084)
	v.SetDefault("server.health.port", 28084)
	v.SetDefault("kafka.consumer_group", "billing-service")
	return commonconfig.Load[ServeConfig](validate, v, "notopia_billing", "billing.notopia.config")
}

var (
	Provide              = New
	ProvideMigrateConfig = NewMigrateConfig
	ProvideSeedConfig    = NewSeedConfig
)

// NewDefaultDomainEvent serves binaries that write billing's outbox without loading billing's config (seed, fixtures).
func NewDefaultDomainEvent() (*DomainEvent, error) {
	cfg := &DomainEvent{}
	if err := defaults.Set(cfg); err != nil {
		return nil, fmt.Errorf("cannot set default domain event config: %w", err)
	}
	return cfg, nil
}

var ProvideDefaultDomainEvent = NewDefaultDomainEvent
