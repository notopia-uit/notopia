package config

import (
	"fmt"
	"log/slog"
	"strings"

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
	commonconfig.Meilisearch
	NoteSearchKeyUID         string `json:"noteSearchKeyUid"         mapstructure:"note_search_key_uid"         validate:"required" yaml:"note_search_key_uid"`
	NoteIndexName            string `json:"noteIndexName"            mapstructure:"note_index_name"             validate:"required" yaml:"note_index_name"`
	NoteSearchExpireDuration int    `json:"noteSearchExpireDuration" mapstructure:"note_search_expire_duration" validate:"required" yaml:"note_search_expire_duration"`
}

func setViperMeilisearchDefault(viper *viper.Viper) {
	viper.SetDefault("meilisearch.note_index_name", "notes")
	viper.SetDefault("meilisearch.note_search_expire_duration", 3600)
}

type DomainEvent struct {
	MessageMetadataUserIDKey      string `json:"messageMetadataUserIdKey"      mapstructure:"message_metadata_user_id_key"      validate:"required" yaml:"message_metadata_user_id_key"`
	MessageWorkspaceIDKey         string `json:"messageMetadataWorkspaceIdKey" mapstructure:"message_metadata_workspace_id_key" validate:"required" yaml:"message_metadata_workspace_id_key"`
	MessageMetadataAggregateIDKey string `json:"messageMetadataAggregateIdKey" mapstructure:"message_metadata_aggregate_id_key" validate:"required" yaml:"message_metadata_aggregate_id_key"`
	OutboxTableName               string `json:"outboxTableName"               mapstructure:"outbox_table_name"                 validate:"required" yaml:"outbox_table_name"`
}

func setViperAdvancedDomainEventDefault(viper *viper.Viper) {
	viper.SetDefault("advanced.domain_event.message_metadata_user_id_key", "user_id")
	viper.SetDefault("advanced.domain_event.message_metadata_workspace_id_key", "workspace_id")
	viper.SetDefault("advanced.domain_event.message_metadata_aggregate_id_key", "aggregate_id")
	viper.SetDefault("advanced.domain_event.outbox_table_name", "eventsToForward")
}

type WorkspaceEvent struct {
	MessageMetadataWorkspaceIDKey string `json:"messageMetadataWorkspaceIdKey" mapstructure:"message_metadata_workspace_id_key" validate:"required" yaml:"message_metadata_workspace_id_key"`
	MessageMetadataUserIDKey      string `json:"messageMetadataUserIdKey"      mapstructure:"message_metadata_user_id_key"      validate:"required" yaml:"message_metadata_user_id_key"`
	MessageMetadataEventTypeKey   string `json:"messageMetadataEventTypeKey"   mapstructure:"message_metadata_event_type_key"   validate:"required" yaml:"message_metadata_event_type_key"`
	MessageGeneralTopic           string `json:"messageGeneralTopic"           mapstructure:"message_general_topic"             validate:"required" yaml:"message_general_topic"`
}

func setViperAdvancedWorkspaceEventDefault(viper *viper.Viper) {
	viper.SetDefault("advanced.workspace_event.message_metadata_workspace_id_key", "workspace_id")
	viper.SetDefault("advanced.workspace_event.message_metadata_user_id_key", "user_id")
	viper.SetDefault("advanced.workspace_event.message_metadata_event_type_key", "event_type")
	viper.SetDefault("advanced.workspace_event.message_general_topic", "events:workspaces")
}

type Advanced struct {
	DomainEvent    DomainEvent    `json:"domainEvent"    mapstructure:"domain_event"    validate:"omitempty" yaml:"domain_event"`
	WorkspaceEvent WorkspaceEvent `json:"workspaceEvent" mapstructure:"workspace_event" validate:"omitempty" yaml:"workspace_event"`
}

func setViperAdvancedDefault(viper *viper.Viper) {
	setViperAdvancedDomainEventDefault(viper)
	setViperAdvancedWorkspaceEventDefault(viper)
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

	commonconfig.ServerAddressViperSetDefault(viper, "server.http", 8081)
	commonconfig.ServerAddressViperSetDefault(viper, "server.grpc", 18081)
	commonconfig.ServerAddressViperSetDefault(viper, "server.health", 28081)
	setViperAdvancedDefault(viper)
	commonconfig.LogViperSetDefault(viper, "log")
	commonconfig.KafkaViperSetDefault(viper, "kafka", "note-service")
	commonconfig.SQLViperSetDefault(viper, "database")
	commonconfig.GeneralViperSetDefault(viper, "general")
	setViperMeilisearchDefault(viper)
	commonconfig.AuthentikViperSetDefault(viper, "authentik")

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		slog.Info("configuration loaded", slog.String("file", viper.ConfigFileUsed()))
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config from env or config file: %w", err)
	}

	slog.Info("configuration", slog.Any("config", cfg))

	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("Config validation failed: %w", err)
	}

	return &cfg, nil
}

var Provide = New
