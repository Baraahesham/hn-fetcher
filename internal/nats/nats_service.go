package nats

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

type NatsClient struct {
	natsConn *nats.Conn
	logger   zerolog.Logger
}
type NewNatsClientParams struct {
	Logger  *zerolog.Logger
	NatsUrl string
}

func New(params NewNatsClientParams) (*NatsClient, error) {
	logger := params.Logger.With().Str("component", "NatsClient").Logger()
	natsConn, err := nats.Connect(params.NatsUrl)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to NATS")
		return nil, err
	}
	return &NatsClient{
		natsConn: natsConn,
		logger:   logger,
	}, nil
}

func (client *NatsClient) Publish(subject string, msg any) error {
	client.logger.Info().Str("subject", subject).Msg("Publishing message to NATS")
	client.logger.Info().Interface("message", msg).Msg("Message content")
	data, err := json.Marshal(msg)
	if err != nil {
		client.logger.Error().Err(err).Msg("Failed to marshal message")
		return err
	}
	return client.natsConn.Publish(subject, data)
}

func (client *NatsClient) Close() {
	if client.natsConn != nil {
		client.logger.Info().Msg("Closing NATS connection")
		client.natsConn.Close()
	}
}
