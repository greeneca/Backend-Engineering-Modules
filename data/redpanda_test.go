package data

import (
	"context"
	"sync"
	"testing"
	"wiki_updates/models"

	"github.com/stretchr/testify/assert"
)

// fakeStreamer is a hand-rolled StreamerInterface for deterministic tests: it
// records everything produced and never blocks (Poll returns no records).
type fakeStreamer struct {
	mu       sync.Mutex
	produced []StreamRecord
	produceErr error
}

func (f *fakeStreamer) Produce(topic string, key, value []byte) error {
	if f.produceErr != nil {
		return f.produceErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.produced = append(f.produced, StreamRecord{Topic: topic, Key: key, Value: value})
	return nil
}

func (f *fakeStreamer) Poll(ctx context.Context) ([]StreamRecord, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (f *fakeStreamer) Close() {}

// newTestRedpanda builds a Redpanda with an injected streamer and an initialized
// materialized view, without starting the background consumer goroutine.
func newTestRedpanda(streamer StreamerInterface) *Redpanda {
	return &Redpanda{
		streamer:      streamer,
		updatesTopic:  "wiki.updates",
		usersTopic:    "wiki.users",
		snapshotTopic: "wiki.stats.snapshot",
		urls:          make(map[string]struct{}),
		bots:          make(map[string]struct{}),
		nonBots:       make(map[string]struct{}),
		users:         make(map[string]models.User),
	}
}

func Test_Redpanda_SaveUpdate_ProducesToUpdatesTopic(t *testing.T) {
	streamer := &fakeStreamer{}
	db := newTestRedpanda(streamer)
	update := models.Update{Uri: "https://en.wikipedia.org/wiki/Go", User: "Gopher", Bot: true}

	err := db.SaveUpdate(update)

	assert.NoError(t, err)
	assert.Len(t, streamer.produced, 1)
	record := streamer.produced[0]
	assert.Equal(t, "wiki.updates", record.Topic)
	assert.Equal(t, update.Uri, string(record.Key))

	decoded, err := decodeUpdate(record.Value)
	assert.NoError(t, err)
	assert.Equal(t, update, decoded)
}

func Test_Redpanda_ExactDistinctCounts(t *testing.T) {
	db := newTestRedpanda(&fakeStreamer{})

	updates := []models.Update{
		{Uri: "uri-a", User: "bot-1", Bot: true},
		{Uri: "uri-a", User: "bot-1", Bot: true},   // duplicate uri + bot
		{Uri: "uri-b", User: "bot-2", Bot: true},   // new uri + bot
		{Uri: "uri-c", User: "human-1", Bot: false}, // new uri + non-bot
		{Uri: "uri-c", User: "human-1", Bot: false}, // duplicate
		{Uri: "uri-d", User: "human-2", Bot: false}, // new uri + non-bot
	}
	for i, update := range updates {
		db.applyUpdate(update, int64(i))
	}

	stats, err := db.GetStatistics()
	assert.NoError(t, err)
	assert.Equal(t, 6, stats.Messages, "every update counts as a message")
	assert.Equal(t, 4, stats.Urls, "distinct uris: a,b,c,d")
	assert.Equal(t, 2, stats.Bots, "distinct bots: bot-1,bot-2")
	assert.Equal(t, 2, stats.NonBots, "distinct non-bots: human-1,human-2")
}

func Test_Redpanda_SaveUpdate_ThenConsume_MaterializesExactCounts(t *testing.T) {
	streamer := &fakeStreamer{}
	db := newTestRedpanda(streamer)

	updates := []models.Update{
		{Uri: "uri-a", User: "bot-1", Bot: true},
		{Uri: "uri-a", User: "bot-1", Bot: true},
		{Uri: "uri-b", User: "human-1", Bot: false},
	}
	for _, update := range updates {
		assert.NoError(t, db.SaveUpdate(update))
	}
	// Simulate the consumer reading the produced records back off the log.
	for _, record := range streamer.produced {
		db.apply(record)
	}

	stats, _ := db.GetStatistics()
	assert.Equal(t, 3, stats.Messages)
	assert.Equal(t, 2, stats.Urls)
	assert.Equal(t, 1, stats.Bots)
	assert.Equal(t, 1, stats.NonBots)
}

func Test_Redpanda_SaveUser_ProducesAndIsImmediatelyReadable(t *testing.T) {
	streamer := &fakeStreamer{}
	db := newTestRedpanda(streamer)
	user := &models.User{Email: "test@user.com", PasswordHash: "hashed"}

	err := db.SaveUser(user)
	assert.NoError(t, err)

	assert.Len(t, streamer.produced, 1)
	assert.Equal(t, "wiki.users", streamer.produced[0].Topic)
	assert.Equal(t, user.Email, string(streamer.produced[0].Key))

	got, err := db.GetUserByEmail(user.Email)
	assert.NoError(t, err)
	assert.Equal(t, user.Email, got.Email)
	assert.Equal(t, user.PasswordHash, got.PasswordHash)
}

func Test_Redpanda_SaveUser_DuplicateReturnsError(t *testing.T) {
	db := newTestRedpanda(&fakeStreamer{})
	user := &models.User{Email: "test@user.com", PasswordHash: "hashed"}

	assert.NoError(t, db.SaveUser(user))
	err := db.SaveUser(user)
	assert.Error(t, err)
}

func Test_Redpanda_GetUserByEmail_NotFound(t *testing.T) {
	db := newTestRedpanda(&fakeStreamer{})
	_, err := db.GetUserByEmail("missing@user.com")
	assert.Error(t, err)
}

func Test_Redpanda_Apply_UserRecordMaterializesUser(t *testing.T) {
	db := newTestRedpanda(&fakeStreamer{})
	user := models.User{Email: "streamed@user.com", PasswordHash: "hashed"}
	value, _ := encodeUser(user)

	db.apply(StreamRecord{Topic: "wiki.users", Key: []byte(user.Email), Value: value})

	got, err := db.GetUserByEmail(user.Email)
	assert.NoError(t, err)
	assert.Equal(t, user.Email, got.Email)
}

func Test_Redpanda_Apply_TracksUpdatesOffset(t *testing.T) {
	db := newTestRedpanda(&fakeStreamer{})
	update := models.Update{Uri: "uri-a", User: "bot-1", Bot: true}
	value, _ := encodeUpdate(update)

	db.apply(StreamRecord{Topic: "wiki.updates", Offset: 41, Value: value})

	// Next offset to consume is one past the last applied record.
	assert.Equal(t, int64(42), db.buildSnapshot().Offset)
}

func Test_Redpanda_SnapshotRoundTrip(t *testing.T) {
	source := newTestRedpanda(&fakeStreamer{})
	updates := []models.Update{
		{Uri: "uri-a", User: "bot-1", Bot: true},
		{Uri: "uri-b", User: "human-1", Bot: false},
		{Uri: "uri-a", User: "bot-1", Bot: true}, // duplicate
	}
	for i, update := range updates {
		source.applyUpdate(update, int64(i))
	}

	snapshot := source.buildSnapshot()
	assert.Equal(t, int64(3), snapshot.Offset)

	// Restore into a fresh instance and confirm the view is identical.
	restored := newTestRedpanda(&fakeStreamer{})
	restored.restoreSnapshot(snapshot)

	stats, _ := restored.GetStatistics()
	assert.Equal(t, 3, stats.Messages)
	assert.Equal(t, 2, stats.Urls)
	assert.Equal(t, 1, stats.Bots)
	assert.Equal(t, 1, stats.NonBots)
	assert.Equal(t, int64(3), restored.buildSnapshot().Offset)
}

func Test_Redpanda_WriteSnapshot_ProducesToSnapshotTopic(t *testing.T) {
	streamer := &fakeStreamer{}
	db := newTestRedpanda(streamer)
	db.applyUpdate(models.Update{Uri: "uri-a", User: "bot-1", Bot: true}, 7)

	assert.NoError(t, db.writeSnapshot())

	assert.Len(t, streamer.produced, 1)
	record := streamer.produced[0]
	assert.Equal(t, "wiki.stats.snapshot", record.Topic)
	assert.Equal(t, snapshotKey, string(record.Key))

	decoded, err := decodeSnapshot(record.Value)
	assert.NoError(t, err)
	assert.Equal(t, 1, decoded.Messages)
	assert.Equal(t, int64(8), decoded.Offset)
	assert.ElementsMatch(t, []string{"uri-a"}, decoded.Urls)
}

func Test_Redpanda_Recover_ThenResumeWithoutDoubleCounting(t *testing.T) {
	// A snapshot captured 5 messages and 2 distinct uris, up to offset 5.
	base := statsSnapshot{
		Messages: 5,
		Urls:     []string{"uri-a", "uri-b"},
		Bots:     []string{"bot-1"},
		NonBots:  []string{"human-1"},
		Offset:   5,
	}
	db := newTestRedpanda(&fakeStreamer{})
	db.restoreSnapshot(base)

	// Resume consuming updates at offset 5. "uri-a" is a replay of a URI already
	// in the snapshot set, so it must not inflate the distinct URI count.
	resumed := []StreamRecord{
		record(t, "wiki.updates", 5, models.Update{Uri: "uri-a", User: "bot-1", Bot: true}),
		record(t, "wiki.updates", 6, models.Update{Uri: "uri-c", User: "human-2", Bot: false}),
	}
	for _, rec := range resumed {
		db.apply(rec)
	}

	stats, _ := db.GetStatistics()
	assert.Equal(t, 7, stats.Messages, "5 base + 2 resumed")
	assert.Equal(t, 3, stats.Urls, "uri-a,uri-b,uri-c (uri-a not double counted)")
	assert.Equal(t, 1, stats.Bots, "bot-1 already known")
	assert.Equal(t, 2, stats.NonBots, "human-1,human-2")
	assert.Equal(t, int64(7), db.buildSnapshot().Offset)
}

func record(t *testing.T, topic string, offset int64, update models.Update) StreamRecord {
	t.Helper()
	value, err := encodeUpdate(update)
	assert.NoError(t, err)
	return StreamRecord{Topic: topic, Offset: offset, Key: []byte(update.Uri), Value: value}
}
