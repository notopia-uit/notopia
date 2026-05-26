package config

import (
	"fmt"
	"log/slog"
	"strings"
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

type Meilisearch struct {
	commonconfig.Meilisearch `default:""      json:""                         mapstructure:",squash"                     validate:"required"               yaml:""`
	NoteSearchKeyUID         string        `default:"-"     json:"noteSearchKeyUid"         mapstructure:"note_search_key_uid"         validate:"required,uuid4_rfc4122" yaml:"note_search_key_uid"`
	NoteIndexName            string        `default:"notes" json:"noteIndexName"            mapstructure:"note_index_name"             validate:"required"               yaml:"note_index_name"`
	NoteSearchExpireDuration time.Duration `default:"1h"    json:"noteSearchExpireDuration" mapstructure:"note_search_expire_duration" validate:"required,gte=0"         yaml:"note_search_expire_duration"`
}

type DomainEvent struct {
	MessageMetadataUserIDKey      string `default:"user_id"         json:"messageMetadataUserIdKey"      mapstructure:"message_metadata_user_id_key"      validate:"required" yaml:"message_metadata_user_id_key"`
	MessageWorkspaceIDKey         string `default:"workspace_id"    json:"messageMetadataWorkspaceIdKey" mapstructure:"message_metadata_workspace_id_key" validate:"required" yaml:"message_metadata_workspace_id_key"`
	MessageMetadataAggregateIDKey string `default:"aggregate_id"    json:"messageMetadataAggregateIdKey" mapstructure:"message_metadata_aggregate_id_key" validate:"required" yaml:"message_metadata_aggregate_id_key"`
	OutboxTableName               string `default:"eventsToForward" json:"outboxTableName"               mapstructure:"outbox_table_name"                 validate:"required" yaml:"outbox_table_name"`
}

type WorkspaceEvent struct {
	MessageMetadataWorkspaceIDKey string `default:"workspace_id"      json:"messageMetadataWorkspaceIdKey" mapstructure:"message_metadata_workspace_id_key" validate:"required" yaml:"message_metadata_workspace_id_key"`
	MessageMetadataUserIDKey      string `default:"user_id"           json:"messageMetadataUserIdKey"      mapstructure:"message_metadata_user_id_key"      validate:"required" yaml:"message_metadata_user_id_key"`
	MessageMetadataEventTypeKey   string `default:"event_type"        json:"messageMetadataEventTypeKey"   mapstructure:"message_metadata_event_type_key"   validate:"required" yaml:"message_metadata_event_type_key"`
	MessageGeneralTopic           string `default:"events:workspaces" json:"messageGeneralTopic"           mapstructure:"message_general_topic"             validate:"required" yaml:"message_general_topic"`
}

type Advanced struct {
	DomainEvent    DomainEvent    `json:"domainEvent"    mapstructure:"domain_event"    validate:"omitempty" yaml:"domain_event"`
	WorkspaceEvent WorkspaceEvent `json:"workspaceEvent" mapstructure:"workspace_event" validate:"omitempty" yaml:"workspace_event"`
}

type Config struct {
	General     commonconfig.General   `json:"general"     mapstructure:"general"     validate:"omitempty" yaml:"general"`
	Log         commonconfig.Log       `json:"log"         mapstructure:"log"         validate:"omitempty" yaml:"log"`
	Server      Server                 `json:"server"      mapstructure:"server"      validate:"required"  yaml:"server"`
	Database    commonconfig.SQL       `json:"database"    mapstructure:"database"    validate:"required"  yaml:"database"`
	Kafka       commonconfig.Kafka     `json:"kafka"       mapstructure:"kafka"       validate:"required"  yaml:"kafka"`
	Redis       commonconfig.Redis     `json:"redis"       mapstructure:"redis"       validate:"required"  yaml:"redis"`
	Authentik   commonconfig.Authentik `json:"authentik"   mapstructure:"authentik"   validate:"required"  yaml:"authentik"`
	Meilisearch Meilisearch            `json:"meilisearch" mapstructure:"meilisearch" validate:"required"  yaml:"meilisearch"`
	Services    Services               `json:"services"    mapstructure:"services"    validate:"required"  yaml:"services"`
	Advanced    Advanced               `json:"advanced"    mapstructure:"advanced"    validate:"omitempty" yaml:"advanced"`
}

func New(
	validate *validator.Validate,
	viper *viper.Viper,
) (*Config, error) {
	viper.SetEnvPrefix("notopia_note")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetConfigName("note.notopia.config")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		slog.Info("configuration loaded", slog.String("file", viper.ConfigFileUsed()))
	}

	cfg := &Config{}
	if err := defaults.Set(cfg); err != nil {
		return nil, fmt.Errorf("cannot set default config: %w", err)
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config from env or config file: %w", err)
	}

	if cfg.Server.HTTP.Port == 0 {
		cfg.Server.HTTP.Port = 8081
	}
	if cfg.Server.GRPC.Port == 0 {
		cfg.Server.GRPC.Port = 18081
	}
	if cfg.Server.Health.Port == 0 {
		cfg.Server.Health.Port = 28081
	}

	if cfg.Kafka.ConsumerGroup == "" {
		cfg.Kafka.ConsumerGroup = "note-service"
	}

	slog.Info("configuration", slog.Any("config", cfg))

	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("Config validation failed: %w", err)
	}

	return cfg, nil
}

var Provide = New
