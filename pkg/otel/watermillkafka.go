package otel

import (
	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/dnwe/otelsarama"
	"go.opentelemetry.io/otel/sdk/trace"
)

type WatermillKafkaTracer struct {
	tp *trace.TracerProvider
}

func NewOTELSaramaTracer(tp *trace.TracerProvider) *WatermillKafkaTracer {
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
