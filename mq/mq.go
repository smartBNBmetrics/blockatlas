package mq

import (
	"github.com/streadway/amqp"
	"github.com/trustwallet/blockatlas/pkg/logger"
	"github.com/trustwallet/blockatlas/storage"
)

var (
	amqpChan *amqp.Channel
	conn     *amqp.Connection
	queue    amqp.Queue
)

type (
	Queue    string
	Consumer func(amqp.Delivery, storage.Addresses)
)

const (
	Transactions  Queue = "transactions"
	Subscriptions Queue = "subscriptions"
)

func Init(uri string) (err error) {
	conn, err = amqp.Dial(uri)
	if err != nil {
		return
	}
	amqpChan, err = conn.Channel()
	if err != nil {
		return
	}
	return
}

func Close() {
	amqpChan.Close()
	conn.Close()
}

func (q Queue) Declare() error {
	_, err := amqpChan.QueueDeclare(string(q), true, false, false, false, nil)
	return err
}

func (q Queue) Publish(body []byte) error {
	return amqpChan.Publish("", string(q), false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         body,
	})
}

func (q Queue) RunConsumer(consumer Consumer, cache storage.Addresses) {
	messageChannel, err := amqpChan.Consume(
		string(q),
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Error(err)
		return
	}

	for data := range messageChannel {
		consumer(data, cache)
	}
}
