package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/utilitywarehouse/go-pubsub"
)

func status(brokerAddrs []string, topic string) (*pubsub.Status, error) {
	status := &pubsub.Status{}

	client, err := sarama.NewClient(brokerAddrs, sarama.NewConfig())
	if err != nil {
		return nil, err
	}
	defer client.Close()

	writablePartitions, err := client.WritablePartitions(topic)
	if err != nil {
		status.Working = false
		status.Problems = append(status.Problems, err.Error())
		return status, nil
	}
	if len(writablePartitions) == 0 {
		status.Working = false
		status.Problems = append(status.Problems, "no writable partitions")
		return status, nil
	}

	status.Working = true
	return status, nil
}
