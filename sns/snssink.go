package sns

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/denislobanov/go-pubsub"
	"github.com/pkg/errors"
)

type messageSink struct {
	client snsiface.SNSAPI
	topic  string
}

// NewSNSSink is a constructor for new AWS SNS MessageSink type
func NewSNSSink(conf *aws.Config, topic string) (pubsub.MessageSink, error) {
	sess, err := session.NewSession(conf)
	if err != nil {
		return nil, errors.Wrap(err, "could not construct AWS Session")
	}

	return &messageSink{
		client: sns.New(sess),
		topic:  topic,
	}, nil
}

// PutMessage sends ProducerMessage types to AWS SNS
func (s *messageSink) PutMessage(message pubsub.ProducerMessage) error {
	b, err := message.Marshal()
	if err != nil {
		return errors.Wrap(err, "SNS Sink could not marshal ProducerMessage")
	}

	input := &sns.PublishInput{
		Message:  aws.String(string(b)),
		TopicArn: &s.topic,
	}

	_, err = s.client.Publish(input)
	if err != nil {
		return errors.Wrap(err, "error publishing to SNS")
	}

	return nil
}

// Status used to check status of connection to AWS SNS
func (s *messageSink) Status() (*pubsub.Status, error) {
	topics, err := s.client.ListTopics(&sns.ListTopicsInput{})
	if err != nil {
		return nil, err
	}

	status := &pubsub.Status{}
	if len(topics.Topics) < 1 {
		status.Working = false
		status.Problems = append(status.Problems, "no SNS topics")
		return status, nil
	}

	for _, x := range topics.Topics {
		if *x.TopicArn == s.topic {
			status.Working = true
			return status, nil
		}
	}

	status.Working = false
	status.Problems = append(status.Problems, fmt.Sprintf("%s not in topic list [%v]", s.topic, topics.Topics))
	return status, nil
}

// Close is stubbed as there is no equivalent on the SNS Client.
func (s *messageSink) Close() error {
	return nil
}
