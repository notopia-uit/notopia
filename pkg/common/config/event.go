package commonconfig

import "github.com/spf13/viper"

type Kafka struct {
	ConsumerGroup string   `json:"consumerGroup" mapstructure:"consumer_group" validate:"required" yaml:"consumer_group"`
	Brokers       []string `json:"brokers"       mapstructure:"brokers"        validate:"required" yaml:"brokers"`
}

func KafkaViperSetDefault(
	viper *viper.Viper,
	prefix string,
	consumerGroupDefault string,
) {
	viper.SetDefault(prefix+".consumer_group", consumerGroupDefault)
}
