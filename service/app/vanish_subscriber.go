package app

import (
	"context"
	"os"
	"time"

	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/redis/go-redis/v9"
)

type VanishSubscriber struct {
	rdb                 *redis.Client
	transactionProvider TransactionProvider
	logger              logging.Logger
}

func NewVanishSubscriber(
	transactionProvider TransactionProvider,
	logger logging.Logger,
) *VanishSubscriber {
	log := logger.New("vanishSubscriber")
	redisURL := os.Getenv("REDIS_URL")

	options, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Error().Message("Error parsing REDIS_URL")
	}

	rdb := redis.NewClient(options)

	return &VanishSubscriber{
		rdb:                 rdb,
		transactionProvider: transactionProvider,
		logger:              log,
	}
}

// Processes messages from the vanish_requests stream and updates the last_id when done
func (f *VanishSubscriber) Run(ctx context.Context) error {
	streamName := "vanish_requests"
	lastProcessedIDKey := "vanish_requests:crossposting_service:last_id"

	lastProcessedID, err := f.rdb.Get(ctx, lastProcessedIDKey).Result()
	if err == redis.Nil {
		lastProcessedID = "0-0"
	} else if err != nil {
		f.logger.Error().Message("Error fetching last processed ID")
		return err
	}

	f.logger.Debug().WithField("lastProcessedID", lastProcessedID).Message("Starting VanishSubscriber")

	for {
		select {
		case <-ctx.Done():
			f.logger.Debug().Message("context canceled, shutting down VanishSubscriber")
			return nil

		default:
			streamEntries, err := f.rdb.XRead(ctx, &redis.XReadArgs{
				Streams: []string{streamName, lastProcessedID},
				Count:   1,
				Block:   5 * time.Second,
			}).Result()

			if err == redis.Nil {
				// No new messages in the stream within the block time, continue the loop
				continue
			} else if err != nil {
				f.logger.Error().Message("Error reading from stream")
				return err
			}

			for _, stream := range streamEntries {
				for _, entry := range stream.Messages {
					streamID := entry.ID
					f.logger.Debug().WithField("streamId", streamID).Message("Processing stream ID")

					pubkey, err := domain.NewPublicKeyFromHex(entry.Values["pubkey"].(string))

					if err != nil {
						f.logger.Error().WithError(err).Message("Error parsing pubkey")
						break
					}

					err = f.removePubkeyInfo(ctx, pubkey)
					if err != nil {
						f.logger.Error().WithError(err).WithField("streamId", streamID).Message("Failed to process entry")
						continue
					}

					err = f.rdb.Set(ctx, lastProcessedIDKey, streamID, 0).Err()
					if err != nil {
						f.logger.Error().WithError(err).Message("Error saving last processed ID")
						return err
					}

					lastProcessedID = streamID
					f.logger.Debug().WithField("lastProcessedID", lastProcessedID).Message("Updated last processed ID")
				}
			}
		}
	}
}

// Deletes the public key
func (f *VanishSubscriber) removePubkeyInfo(ctx context.Context, pubkey domain.PublicKey) error {
	err := f.transactionProvider.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		return adapters.PublicKeys.DeleteByPublicKey(pubkey)
	})

	if err != nil {
		f.logger.Error().WithError(err).WithField("pubkey", pubkey).Message("Failed to remove pubkey info")
		return err
	}

	f.logger.Debug().WithField("pubkey", pubkey).Message("Removed pubkey info")

	return nil
}
