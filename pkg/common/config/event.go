package commonconfig

type Kafka struct {
	ConsumerGroup string   `json:"consumerGroup" mapstructure:"consumer_group" validate:"required" yaml:"consumer_group"`
	Brokers       []string `json:"brokers"       mapstructure:"brokers"        validate:"required" yaml:"brokers"`
}
