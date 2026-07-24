package data

import (
	"errors"
	"fmt"
	"wiki_updates/configuration"
	"wiki_updates/data/stores"
	"wiki_updates/models"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// Counter names tracked in the stats counter table.
const (
	statMessages = "messages"
	statUrls     = "urls"
	statBots     = "bots"
	statNonBots  = "non_bots"
)

type Cassandra struct {
	session stores.SessionInterface
	stats models.Statistics
}

func (db *Cassandra) Initialize(config configuration.Config){
	cluster := gocql.NewCluster(config.ClusterHosts()...)
	if config.Debug() {
		cluster.Logger = gocql.NewLogger(gocql.LogLevelDebug)
	}
	cluster.Keyspace = config.ClusterKeyspace()
	session, err := cluster.CreateSession()
	if err != nil {
		fmt.Println("Error creating Cassandra session:", err)
		panic(err)
	}
	db.session = stores.NewSession(session)

	createTables(db.session)

	db.stats = models.Statistics{
		Messages: 0,
		Urls:     0,
		Bots:     0,
		NonBots:  0,
	}
}

func createTables(session stores.SessionInterface) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		email TEXT PRIMARY KEY,
		password_hash TEXT
	)`
	if err := session.Query(query).Exec(); err != nil {
		panic(err)
	}
	// wiki_users holds one row per distinct (bot, user); the IF NOT EXISTS on
	// insert lets us detect newly-seen users so distinct counts stay accurate.
	query = `
	CREATE TABLE IF NOT EXISTS wiki_users (
		bot BOOLEAN,
		user TEXT,
		PRIMARY KEY ((bot), user)
	)`
	if err := session.Query(query).Exec(); err != nil {
		panic(err)
	}
	// uris holds one row per distinct URI.
	query = `
	CREATE TABLE IF NOT EXISTS uris (
		uri TEXT PRIMARY KEY
	)`
	if err := session.Query(query).Exec(); err != nil {
		panic(err)
	}
	// stats keeps running counters so statistics are O(1) point reads instead
	// of full-table COUNT(*) scans, which are a Cassandra anti-pattern.
	query = `
	CREATE TABLE IF NOT EXISTS stats (
		name TEXT PRIMARY KEY,
		value COUNTER
	)`
	if err := session.Query(query).Exec(); err != nil {
		panic(err)
	}
}

func (db *Cassandra) SaveUpdate(update models.Update) error {
	// Every update is a message.
	if err := db.incrementStat(statMessages, 1); err != nil {
		return err
	}

	// Count distinct URIs: only bump the counter when the URI is newly inserted.
	uriApplied, err := db.session.Query(
		`INSERT INTO uris (uri) VALUES (?) IF NOT EXISTS`, update.Uri,
	).MapScanCAS(map[string]any{})
	if err != nil {
		fmt.Println("Error inserting URI:", err)
		return err
	}
	if uriApplied {
		if err := db.incrementStat(statUrls, 1); err != nil {
			return err
		}
	}

	// Count distinct users split by bot/non-bot, again only on first sighting.
	userApplied, err := db.session.Query(
		`INSERT INTO wiki_users (bot, user) VALUES (?, ?) IF NOT EXISTS`, update.Bot, update.User,
	).MapScanCAS(map[string]any{})
	if err != nil {
		fmt.Println("Error inserting wiki_user:", err)
		return err
	}
	if userApplied {
		stat := statNonBots
		if update.Bot {
			stat = statBots
		}
		if err := db.incrementStat(stat, 1); err != nil {
			return err
		}
	}
	return nil
}

// incrementStat bumps a named counter in the stats table.
func (db *Cassandra) incrementStat(name string, delta int) error {
	query := `UPDATE stats SET value = value + ? WHERE name = ?`
	if err := db.session.Query(query, delta, name).Exec(); err != nil {
		fmt.Printf("Error incrementing %s counter: %v\n", name, err)
		return err
	}
	return nil
}

func (db *Cassandra) GetStatistics() (*models.Statistics, error) {
	return &models.Statistics{
		Messages: db.readStat(statMessages),
		Urls:     db.readStat(statUrls),
		Bots:     db.readStat(statBots),
		NonBots:  db.readStat(statNonBots),
	}, nil
}

// readStat returns a named counter's value via a point read. A missing row
// simply means nothing has been counted yet, so it reports 0.
func (db *Cassandra) readStat(name string) int {
	var value int
	query := `SELECT value FROM stats WHERE name = ?`
	if err := db.session.Query(query, name).Scan(&value); err != nil {
		if !errors.Is(err, gocql.ErrNotFound) {
			fmt.Printf("Error reading %s counter: %v\n", name, err)
		}
		return 0
	}
	return value
}

func (db *Cassandra) SaveUser(user *models.User) error {
	storedUser, err := db.GetUserByEmail(user.Email)
	if err == nil && storedUser != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}
	query := `INSERT INTO users (email, password_hash) VALUES (?, ?)`
	if err := db.session.Query(query, user.Email, user.PasswordHash).Exec(); err != nil {
		fmt.Println("Error inserting user:", err)
		return err
	}
	return nil
}

func (db *Cassandra) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT email, password_hash FROM users WHERE email = ? LIMIT 1`
	if err := db.session.Query(query, email).Scan(&user.Email, &user.PasswordHash); err != nil {
		fmt.Println("Error querying user by email:", err)
		return nil, err
	}
	return user, nil
}
