package otel

import (
	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/dnwe/otelsarama"
	wotel "github.com/nkonev/watermill-opentelemetry/pkg/opentelemetry"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func NewWatermillOTELOption(t trace.Tracer) wotel.Option {
	return wotel.WithTracer(t)
}

type WatermillKafkaTracer struct {
	tp *tracesdk.TracerProvider
}

func NewOTELSaramaTracer(tp *tracesdk.TracerProvider) *WatermillKafkaTracer {
	return &WatermillKafkaTracer{tp: tp}
}

var ProvideOTELSaramaTracer = NewOTELSaramaTracer

var _ kafka.SaramaTracer = (*WatermillKafkaTracer)(nil)

func (t *WatermillKafkaTracer) WrapConsumer(c sarama.Consumer) sarama.Consumer {
	return otelsarama.WrapConsumer(c, otelsarama.WithTracerProvider(t.tp))
}

func (t *WatermillKafkaTracer) WrapConsumerGroupHandler(h sarama.ConsumerGroupHandler) sarama.ConsumerGroupHandler {
	return otelsarama.WrapConsumerGroupHandler(h, otelsarama.WithTracerProvider(t.tp))
}

func (t *WatermillKafkaTracer) WrapPartitionConsumer(pc sarama.PartitionConsumer) sarama.PartitionConsumer {
	return otelsarama.WrapPartitionConsumer(pc, otelsarama.WithTracerProvider(t.tp))
}

func (t *WatermillKafkaTracer) WrapSyncProducer(cfg *sarama.Config, p sarama.SyncProducer) sarama.SyncProducer {
	return otelsarama.WrapSyncProducer(cfg, p, otelsarama.WithTracerProvider(t.tp))
}
