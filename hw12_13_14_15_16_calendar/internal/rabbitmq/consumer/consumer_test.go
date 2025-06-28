package consumer

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
)

type ackMock struct{ err error }

func (a *ackMock) Ack(tag uint64, multiple bool) error                { return a.err }
func (a *ackMock) Nack(tag uint64, multiple bool, requeue bool) error { return nil }
func (a *ackMock) Reject(tag uint64, requeue bool) error              { return nil }

func TestAckDeliveryError(t *testing.T) {
	var buf bytes.Buffer
	lg := logger.New("debug", &buf)
	c := &RabbitConsumer{logger: lg}

	d := amqp.Delivery{
		Acknowledger: &ackMock{err: errors.New("ack error")},
		DeliveryTag:  1,
	}

	err := c.ackDelivery(context.Background(), d)
	require.Error(t, err)
	require.Contains(t, buf.String(), "failed to acknowledge message")
}
