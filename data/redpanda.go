package data

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"wiki_updates/configuration"
	"wiki_updates/models"
)

// Indirected constructors so tests can inject fakes / avoid real brokers.
var (
	newConsumerStreamer = NewConsumerStreamer
	newProducerStreamer = NewProducerStreamer
	readLatestSnapshot  = ReadLatestSnapshot
)

const (
	snapshotKey = "global"
	snapshotInterval = 15 * time.Second
)

// statsSnapshot is the serializable representation of the state of the consumed data.
type statsSnapshot struct {
	Messages int
	Urls     []string
	Bots     []string
	NonBots  []string
	Offset   int64
}

type Redpanda struct {
	streamer      StreamerInterface
	updatesTopic  string
	usersTopic    string
	snapshotTopic string
	snapshotEvery time.Duration

	// mu guards the view below.
	mu            sync.Mutex
	messages      int
	urls          map[string]struct{}
	bots          map[string]struct{}
	nonBots       map[string]struct{}
	users         map[string]models.User
	updatesOffset int64 // next updates-topic offset to consume
}

func (r *Redpanda) Initialize(config configuration.Config) {
	r.updatesTopic = config.RedpandaTopic()
	r.usersTopic = config.RedpandaUsersTopic()
	r.snapshotTopic = config.RedpandaSnapshotTopic()
	r.snapshotEvery = snapshotInterval
	r.urls = make(map[string]struct{})
	r.bots = make(map[string]struct{})
	r.nonBots = make(map[string]struct{})
	r.users = make(map[string]models.User)

	// Recover the data from the latest snapshot before consuming,
	// so counters survive restarts and topic retention.
	if raw, err := readLatestSnapshot(config); err != nil {
		fmt.Println("Error reading snapshot:", err)
	} else if raw != nil {
		snapshot, err := decodeSnapshot(raw)
		if err != nil {
			fmt.Println("Error decoding snapshot:", err)
		} else {
			r.restoreSnapshot(snapshot)
			fmt.Printf("Recovered snapshot: %d messages, resuming updates at offset %d\n", snapshot.Messages, snapshot.Offset)
		}
	}

	if r.streamer == nil {
		streamer, err := newConsumerStreamer(config, r.updatesOffset)
		if err != nil {
			fmt.Println("Error creating Redpanda consumer:", err)
			panic(err)
		}
		r.streamer = streamer
	}

	go r.consume()
	go r.snapshotLoop()
}

// SaveUpdate publishes the update to the updates topic keyed by URI. It should be
// unused in the regular redpand build format. It remains here so the monolith mode also works.
func (r *Redpanda) SaveUpdate(update models.Update) error {
	value, err := encodeUpdate(update)
	if err != nil {
		return err
	}
	return r.streamer.Produce(r.updatesTopic, []byte(update.Uri), value)
}

// consume is the aggregation loop that materializes statistics and the user map
// from the log. It runs until the streamer is closed or errors.
func (r *Redpanda) consume() {
	for {
		records, err := r.streamer.Poll(context.Background())
		if err != nil {
			fmt.Println("Error polling Redpanda:", err)
			return
		}
		for _, record := range records {
			r.apply(record)
		}
	}
}

// apply routes a consumed record to the correct materialized-view updater.
func (r *Redpanda) apply(record StreamRecord) {
	switch record.Topic {
	case r.updatesTopic:
		update, err := decodeUpdate(record.Value)
		if err != nil {
			fmt.Println("Error decoding update record:", err)
			return
		}
		r.applyUpdate(update, record.Offset)
	case r.usersTopic:
		user, err := decodeUser(record.Value)
		if err != nil {
			fmt.Println("Error decoding user record:", err)
			return
		}
		r.applyUser(user)
	}
}

// applyUpdate maintains exact distinct counts using seen-sets and advances the
// consumed offset.
func (r *Redpanda) applyUpdate(update models.Update, offset int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages++
	r.urls[update.Uri] = struct{}{}
	if update.Bot {
		r.bots[update.User] = struct{}{}
	} else {
		r.nonBots[update.User] = struct{}{}
	}
	r.updatesOffset = offset + 1
}

func (r *Redpanda) applyUser(user models.User) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.Email] = user
}

func (r *Redpanda) GetStatistics() (*models.Statistics, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return &models.Statistics{
		Messages: r.messages,
		Urls:     len(r.urls),
		Bots:     len(r.bots),
		NonBots:  len(r.nonBots),
	}, nil
}

func (r *Redpanda) SaveUser(user *models.User) error {
	r.mu.Lock()
	_, exists := r.users[user.Email]
	r.mu.Unlock()
	if exists {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	value, err := encodeUser(*user)
	if err != nil {
		return err
	}
	if err := r.streamer.Produce(r.usersTopic, []byte(user.Email), value); err != nil {
		fmt.Println("Error producing user:", err)
		return err
	}

	r.applyUser(*user)
	return nil
}

func (r *Redpanda) GetUserByEmail(email string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, exists := r.users[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

// snapshotLoop periodically persists the materialized view to the snapshot
// topic so state (including exact distinct sets) survives restarts.
func (r *Redpanda) snapshotLoop() {
	ticker := time.NewTicker(r.snapshotEvery)
	defer ticker.Stop()
	for range ticker.C {
		if err := r.writeSnapshot(); err != nil {
			fmt.Println("Error writing snapshot:", err)
		}
	}
}

// writeSnapshot serializes and publishes the current view to the snapshot topic.
func (r *Redpanda) writeSnapshot() error {
	snapshot := r.buildSnapshot()
	value, err := encodeSnapshot(snapshot)
	if err != nil {
		return err
	}
	return r.streamer.Produce(r.snapshotTopic, []byte(snapshotKey), value)
}

// buildSnapshot captures the full materialized view under a single lock.
func (r *Redpanda) buildSnapshot() statsSnapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	return statsSnapshot{
		Messages: r.messages,
		Urls:     setKeys(r.urls),
		Bots:     setKeys(r.bots),
		NonBots:  setKeys(r.nonBots),
		Offset:   r.updatesOffset,
	}
}

// restoreSnapshot replaces the materialized view with a decoded snapshot.
func (r *Redpanda) restoreSnapshot(snapshot statsSnapshot) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = snapshot.Messages
	r.urls = sliceToSet(snapshot.Urls)
	r.bots = sliceToSet(snapshot.Bots)
	r.nonBots = sliceToSet(snapshot.NonBots)
	r.updatesOffset = snapshot.Offset
}

func setKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	return keys
}

func sliceToSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

// Producer publishes wiki updates to the updates topic. It is the ingest half
// of the split deployment (cmd/producer) and holds no materialized state.
type Producer struct {
	streamer StreamerInterface
	topic    string
}

// NewProducer builds a produce-only Redpanda client for the ingest process.
func NewProducer(config configuration.Config) *Producer {
	streamer, err := newProducerStreamer(config)
	if err != nil {
		fmt.Println("Error creating Redpanda producer:", err)
		panic(err)
	}
	return &Producer{streamer: streamer, topic: config.RedpandaTopic()}
}

// Produce publishes a single update keyed by URI. Errors are logged rather than
// returned so a single bad record never tears down the ingest loop.
func (p *Producer) Produce(update models.Update) {
	value, err := encodeUpdate(update)
	if err != nil {
		fmt.Println("Error encoding update:", err)
		return
	}
	if err := p.streamer.Produce(p.topic, []byte(update.Uri), value); err != nil {
		fmt.Println("Error producing update:", err)
	}
}

func (p *Producer) Close() {
	p.streamer.Close()
}
