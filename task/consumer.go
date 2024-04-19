package task

import (
	"github.com/IBM/sarama"
)

type Consumer struct {
	Task *Task
}

func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		value := msg.Value
		c.Task.Push(value)
	}
	return nil
}
