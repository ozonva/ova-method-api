package queue

import (
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog/log"
)

type kafkaProvider struct {
	brokers  []string
	config   *sarama.Config
	producer sarama.SyncProducer
}

func NewKafkaProvider(brokers []string, config *sarama.Config) Queue {
	return &kafkaProvider{brokers: brokers, config: config}
}

func (kafka *kafkaProvider) Connect() error {
	conn, err := sarama.NewSyncProducer(kafka.brokers, kafka.config)
	if err != nil {
		return err
	}

	kafka.producer = conn
	return nil
}

func (kafka *kafkaProvider) Close() error {
	return kafka.producer.Close()
}

func (kafka *kafkaProvider) Send(queueName string, msg QueueMsg) error {
	bytes, err := msg.MarshalJSON()
	if err != nil {
		return err
	}

	kafkaMsg := &sarama.ProducerMessage{
		Topic: queueName,
		Value: sarama.ByteEncoder(bytes),
	}

	partition, offset, err := kafka.producer.SendMessage(kafkaMsg)
	if err != nil {
		return err
	}

	log.Debug().
		Str("topic", queueName).
		Str("msg", string(bytes)).
		Int32("partition", partition).
		Int64("offset", offset).
		Msg("send")

	return nil
}
