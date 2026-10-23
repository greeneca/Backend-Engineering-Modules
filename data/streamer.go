package data

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"wiki_updates/configuration"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

// snapshotMaxBytes is the largest snapshot record (and topic max.message.bytes)
// allowed. The snapshot holds every distinct URI and user, so it grows over time.
const snapshotMaxBytes = 64 * 1024 * 1024

// StreamRecord is a simplified representation of a consumed record to allow the
// streamer to be mocked in tests without exposing franz-go types.
type StreamRecord struct {
	Topic  string
	Offset int64
	Key    []byte
	Value  []byte
}

// StreamerInterface abstracts the Redpanda client so it can be mocked in tests.
type StreamerInterface interface {
	Produce(topic string, key, value []byte) error
	Poll(ctx context.Context) ([]StreamRecord, error)
	Close()
}

// Streamer is the franz-go backed implementation of StreamerInterface.
type Streamer struct {
	client *kgo.Client
}

// NewProducerStreamer builds a produce-only client and ensures the topics
// exist. Used by the producer binary and anywhere updates are published.
func NewProducerStreamer(config configuration.Config) (StreamerInterface, error) {
	client, err := kgo.NewClient(kgo.SeedBrokers(config.RedpandaBrokers()...))
	if err != nil {
		return nil, err
	}
	if err := ensureTopics(client, config); err != nil {
		fmt.Println("Error ensuring Redpanda topics:", err)
	}
	return &Streamer{client: client}, nil
}

// NewConsumerStreamer builds a client that consumes the updates topic starting
// at updatesStartOffset and the (compacted) users topic from the start, and can
// also produce (used by the consumer to write snapshots).
func NewConsumerStreamer(config configuration.Config, updatesStartOffset int64) (StreamerInterface, error) {
	updatesOffset := kgo.NewOffset().AtStart()
	if updatesStartOffset > 0 {
		updatesOffset = kgo.NewOffset().At(updatesStartOffset)
	}
	client, err := kgo.NewClient(
		kgo.SeedBrokers(config.RedpandaBrokers()...),
		kgo.ProducerBatchMaxBytes(snapshotMaxBytes),
		kgo.FetchMaxBytes(snapshotMaxBytes),
		kgo.FetchMaxPartitionBytes(snapshotMaxBytes),
		kgo.MaxBufferedBytes(2*snapshotMaxBytes),
		kgo.ConsumePartitions(map[string]map[int32]kgo.Offset{
			config.RedpandaTopic():      {0: updatesOffset},
			config.RedpandaUsersTopic(): {0: kgo.NewOffset().AtStart()},
		}),
	)
	if err != nil {
		return nil, err
	}
	if err := ensureTopics(client, config); err != nil {
		fmt.Println("Error ensuring Redpanda topics:", err)
	}
	return &Streamer{client: client}, nil
}

// ensureTopics creates the updates topic (delete/retention policy) and the
// users + snapshot topics (compacted, last-write-wins) if they do not exist.
func ensureTopics(client *kgo.Client, config configuration.Config) error {
	adm := kadm.NewClient(client)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := adm.CreateTopics(ctx, 1, 1, nil, config.RedpandaTopic()); err != nil {
		return err
	}
	compact := "compact"
	compacted := map[string]*string{"cleanup.policy": &compact}
	if _, err := adm.CreateTopics(ctx, 1, 1, compacted, config.RedpandaUsersTopic()); err != nil {
		return err
	}
	maxBytes := strconv.Itoa(snapshotMaxBytes)
	snapshotConfigs := map[string]*string{"cleanup.policy": &compact, "max.message.bytes": &maxBytes}
	if _, err := adm.CreateTopics(ctx, 1, 1, snapshotConfigs, config.RedpandaSnapshotTopic()); err != nil {
		return err
	}
	// CreateTopics does not change a topic that already exists, so also set the
	// size limit explicitly. This fixes snapshot topics created with the default.
	altered, err := adm.AlterTopicConfigs(ctx, []kadm.AlterConfig{
		{Op: kadm.SetConfig, Name: "max.message.bytes", Value: &maxBytes},
	}, config.RedpandaSnapshotTopic())
	if err != nil {
		return err
	}
	for _, result := range altered {
		if result.Err != nil {
			return result.Err
		}
	}
	return nil
}

// ReadLatestSnapshot reads the compacted snapshot topic to its end with a
// short-lived client and returns the most recent record value (nil if empty).
// A dedicated client is used so this one-shot read does not interfere with the
// consumer's live partition assignment.
func ReadLatestSnapshot(config configuration.Config) ([]byte, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(config.RedpandaBrokers()...),
		kgo.FetchMaxBytes(snapshotMaxBytes),
		kgo.FetchMaxPartitionBytes(snapshotMaxBytes),
		kgo.ConsumePartitions(map[string]map[int32]kgo.Offset{
			config.RedpandaSnapshotTopic(): {0: kgo.NewOffset().AtStart()},
		}),
	)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var latest []byte
	for {
		fetches := client.PollFetches(ctx)
		if fetches.IsClientClosed() || len(fetches.Errors()) > 0 {
			break
		}
		count := 0
		fetches.EachRecord(func(r *kgo.Record) {
			latest = r.Value
			count++
		})
		if count == 0 {
			break
		}
	}
	return latest, nil
}

// Produce synchronously publishes a keyed record.
func (s *Streamer) Produce(topic string, key, value []byte) error {
	record := &kgo.Record{Topic: topic, Key: key, Value: value}
	return s.client.ProduceSync(context.Background(), record).FirstErr()
}

// Poll returns the next batch of records fetched from the broker.
func (s *Streamer) Poll(ctx context.Context) ([]StreamRecord, error) {
	fetches := s.client.PollFetches(ctx)
	if fetches.IsClientClosed() {
		return nil, fmt.Errorf("redpanda client closed")
	}
	if errs := fetches.Errors(); len(errs) > 0 {
		return nil, errs[0].Err
	}
	var records []StreamRecord
	fetches.EachRecord(func(r *kgo.Record) {
		records = append(records, StreamRecord{
			Topic:  r.Topic,
			Offset: r.Offset,
			Key:    r.Key,
			Value:  r.Value,
		})
	})
	return records, nil
}

func (s *Streamer) Close() {
	s.client.Close()
}
