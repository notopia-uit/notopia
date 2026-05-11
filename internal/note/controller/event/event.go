package event

import (
	"context"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	wotel "github.com/nkonev/watermill-opentelemetry/pkg/opentelemetry"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/component"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/pkg/api/share"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

// This include integration (from share package) and domain event, setup for event processor
// If not use with event processor but via subscriber directly, not need to declare this
func eventToTopic(event any) (string, error) {
	switch e := event.(type) {
	case domain.Event:
		return component.DomainEventToTopic(e)
	case *share.DocumentCommittedEvent:
		return "events.integration.document.document.committed", nil
	default:
		return "", fmt.Errorf("unknown event type: %T", event)
	}
}

// Be careful if we use single subcriber, they will steal each other message
type Event struct {
	workspaceEventSubcriber message.Subscriber
	subcriber               message.Subscriber
	eventProcessor          *cqrs.EventProcessor
	router                  *message.Router
	app                     *app.Server

	domainEventCfg *config.DomainEvent
}

// I mean, kafka
type Publisher message.Publisher

func NewEvent(
	generalCfg *commonconfig.General,
	kafkaCfg *commonconfig.Kafka,
	app *app.Server,
	tracer kafka.SaramaTracer,
	logger watermill.LoggerAdapter,
	marshaler *cqrs.JSONMarshaler,
	domainEventCfg *config.DomainEvent,
	publisher Publisher,
) (*Event, error) {
	saramaConfig := kafka.DefaultSaramaSubscriberConfig()
	if generalCfg.AppEnv == commonconfig.AppEnvDevelopment {
		// FIXME:
		saramaConfig.Consumer.Group.Session.Timeout = 30 * time.Second
	}
	subcriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               kafkaCfg.Brokers,
			ConsumerGroup:         kafkaCfg.ConsumerGroup,
			Tracer:                tracer,
			OverwriteSaramaConfig: saramaConfig,
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create event controller subscriber: %w", err)
	}
	workspaceEventSubcriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               kafkaCfg.Brokers,
			ConsumerGroup:         kafkaCfg.ConsumerGroup + ".workspace-event",
			Tracer:                tracer,
			OverwriteSaramaConfig: saramaConfig,
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create event controller workspace event subscriber: %w", err)
	}
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create event controller router: %w", err)
	}
	retryMiddleware := middleware.Retry{
		MaxRetries:      3,
		InitialInterval: 500 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      2,
		Logger:          logger,
	}
	poisonQueueMiddleware, err := middleware.PoisonQueue(publisher, component.DomainEventTopicPrefix+"poison")
	if err != nil {
		return nil, fmt.Errorf("failed to create poison queue middleware: %w", err)
	}
	router.AddMiddleware(
		poisonQueueMiddleware,
		middleware.CorrelationID,
		middleware.Recoverer,
		retryMiddleware.Middleware,
		wotel.Trace(),
	)
	eventProcessor, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				topic, err := eventToTopic(params.EventHandler.NewEvent())
				if err != nil {
					return "", fmt.Errorf("failed to get event topic for %s: %w", params.EventName, err)
				}
				return topic, nil
			},
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return kafka.NewSubscriber(
					kafka.SubscriberConfig{
						Brokers:               kafkaCfg.Brokers,
						ConsumerGroup:         kafkaCfg.ConsumerGroup + "." + params.HandlerName,
						Tracer:                tracer,
						OverwriteSaramaConfig: saramaConfig,
					},
					logger,
				)
			},
			Marshaler: marshaler,
			Logger:    logger,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create event controller event processor: %w", err)
	}
	event := &Event{
		workspaceEventSubcriber: workspaceEventSubcriber,
		subcriber:               subcriber,
		eventProcessor:          eventProcessor,
		router:                  router,
		app:                     app,
		domainEventCfg:          domainEventCfg,
	}
	if err := event.setup(); err != nil {
		return nil, fmt.Errorf("failed to setup event controller: %w", err)
	}
	return event, nil
}

var ProvideEvent = NewEvent

func (e *Event) setup() error {
	if _, err := e.eventProcessor.AddHandler(cqrs.NewEventHandler(
		"DocumentCommittedHandler",
		e.documentCommittedHandler,
	)); err != nil {
		return fmt.Errorf("failed to add event handler DocumentCommittedHandler: %w", err)
	}

	if _, err := e.eventProcessor.AddHandler(cqrs.NewEventHandler(
		"NotifyWorkspaceRenamedHandler",
		e.app.Events.NotifyWorkspaceRenamedHandler.Handle,
	)); err != nil {
		return fmt.Errorf("failed to add event handler NotifyWorkspaceRenamedHandler: %w", err)
	}

	if _, err := e.eventProcessor.AddHandler(cqrs.NewEventHandler(
		"NotifyWorkspaceSlugChangedHandler",
		e.app.Events.NotifyWorkspaceSlugChangedHandler.Handle,
	)); err != nil {
		return fmt.Errorf("failed to add event handler NotifyWorkspaceSlugChangedHandler: %w", err)
	}

	// NOTE: because watermill doesn't support kafka regex (IBM/sarama)
	// So, we will need to for loop all topic we have, (for note, and folder)
	workspaceItemUpdatedFolderTopics := []string{
		component.DomainEventTopicPrefix + "folder.created",
		component.DomainEventTopicPrefix + "folder.deleted",
		component.DomainEventTopicPrefix + "folder.renamed",
		component.DomainEventTopicPrefix + "folder.icon_changed",
		component.DomainEventTopicPrefix + "folder.moved",
		component.DomainEventTopicPrefix + "folder.trashed",
		component.DomainEventTopicPrefix + "folder.restored",
		component.DomainEventTopicPrefix + "folder.permanently_deleted",
	}
	for _, topic := range workspaceItemUpdatedFolderTopics {
		e.router.AddConsumerHandler(
			fmt.Sprintf("NotifyWorkspaceItemsUpdatedHandler.%s", topic),
			topic,
			e.workspaceEventSubcriber,
			e.notifyWorkspaceItemsUpdatedFolderHandler,
		)
	}
	workspaceItemUpdatedNoteTopics := []string{
		component.DomainEventTopicPrefix + "note.created",
		component.DomainEventTopicPrefix + "note.deleted",
		component.DomainEventTopicPrefix + "note.renamed",
		component.DomainEventTopicPrefix + "note.icon_changed",
		component.DomainEventTopicPrefix + "note.tags_changed",
		component.DomainEventTopicPrefix + "note.size_changed",
		component.DomainEventTopicPrefix + "note.outgoing_links_changed",
		component.DomainEventTopicPrefix + "note.moved",
		component.DomainEventTopicPrefix + "note.trashed",
		component.DomainEventTopicPrefix + "note.restored",
		component.DomainEventTopicPrefix + "note.permanently_deleted",
	}
	for _, topic := range workspaceItemUpdatedNoteTopics {
		e.router.AddConsumerHandler(
			fmt.Sprintf("NotifyWorkspaceItemsUpdatedHandler.%s", topic),
			topic,
			e.workspaceEventSubcriber,
			e.notifyWorkspaceItemsUpdatedNoteHandler,
		)
	}
	noteUpdatedTopics := []string{
		component.DomainEventTopicPrefix + "note.renamed",
		component.DomainEventTopicPrefix + "note.icon_changed",
		component.DomainEventTopicPrefix + "note.moved",
		component.DomainEventTopicPrefix + "note.trashed",
		component.DomainEventTopicPrefix + "note.restored",
	}
	for _, topic := range noteUpdatedTopics {
		e.router.AddConsumerHandler(
			fmt.Sprintf("NoteUpdatedHandler.%s", topic),
			topic,
			e.subcriber,
			e.noteUpdatedToIntegrationHandler,
		)
	}
	return nil
}

func (e *Event) Run(ctx context.Context) error {
	if err := e.router.Run(ctx); err != nil {
		return fmt.Errorf("failed to run event controller router: %w", err)
	}
	return nil
}

func (e *Event) IsRunning() bool {
	return e.router.IsRunning()
}
