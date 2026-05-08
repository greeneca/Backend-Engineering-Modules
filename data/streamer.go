package data

import (
	"context"
	"encoding/json"
	"fmt"
	"wiki_updates/configuration"
	"wiki_updates/models"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

type DataStreamer struct {
	client *kgo.Client
	topic string
}

func GetDataStreamer(config configuration.Config) DataStreamer {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(config.KafkaBrokers()...),
		kgo.ConsumerGroup(config.KafkaConsumerGroup()),
		kgo.ConsumeTopics(config.KafkaTopic()),
	)
	if err != nil {
		fmt.Println("Error creating Kafka client:", err)
		panic(err)
	}

	// Create the topic if it doesn't exist
	adm := kadm.NewClient(client)
	_, err = adm.CreateTopics(context.Background(), 1, 1, nil, config.KafkaTopic())
	if err != nil {
		fmt.Println("Error creating Kafka topic:", err)
		panic(err)
	}

	return DataStreamer{
		client: client,
		topic: config.KafkaTopic(),
	}
}

func (ds *DataStreamer) Produce(update models.Update) {
	data, _ := json.Marshal(update)
	msg := &kgo.Record{
		Topic: ds.topic,
		Value: []byte(data),
	}
	results := ds.client.ProduceSync(context.Background(), msg)
	err := results.FirstErr()
	if err != nil {
		fmt.Println("Error producing message to Kafka:", err)
	}
}

func (ds *DataStreamer) Consume(process func(models.Update)) error {
	for {
		fetches := ds.client.PollFetches(context.Background())
		if fetches.IsClientClosed() {
			return fmt.Errorf("Kafka client is closed")
		}
		if fetches.NumRecords() == 0 {
			return nil // No messages available
		}
		iter := fetches.RecordIter()
		fmt.Printf("Processing %d messages from Kafka\n", fetches.NumRecords())
		for !iter.Done() {
			var update models.Update
			record := iter.Next()
			err := json.Unmarshal(record.Value, &update)
			if err != nil {
				fmt.Println("Error unmarshalling Kafka message:", err, "message:", string(record.Value))
				continue
			}
			process(update)
		}
	}
}

func (ds *DataStreamer) Close() {
	ds.client.Close()
}
