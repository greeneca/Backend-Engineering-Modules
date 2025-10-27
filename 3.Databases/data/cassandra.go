package data

import (
	"fmt"
	"wiki_updates/configuration"
	"wiki_updates/models"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

type Cassandra struct {
	Session *gocql.Session
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
	db.Session = session

	createTables(session)

	db.stats = models.Statistics{
		Messages: 0,
		Urls:     0,
		Bots:     0,
		NonBots:  0,
	}
}

func createTables(session *gocql.Session) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		email TEXT,
		password_hash TEXT,
		created_at TIMESTAMP
	)`
	if err := session.Query(query).Exec(); err != nil {
		panic(err)
	}
	query = `
	CREATE TABLE IF NOT EXISTS wiki_users (
		id UUID,
		user TEXT,
		bot BOOLEAN,
		PRIMARY KEY ((bot), user, id)
	)`
	if err := session.Query(query).Exec(); err != nil {
		panic(err)
	}
	query = `
	CREATE TABLE IF NOT EXISTS uris (
		id UUID,
		uri TEXT PRIMARY KEY
	)`
	if err := session.Query(query).Exec(); err != nil {
		panic(err)
	}
	query = `
	CREATE TABLE IF NOT EXISTS updates (
		id UUID,
		uri_id UUID,
		user_id UUID,
		PRIMARY KEY ((user_id), uri_id, id)
	)`
	if err := session.Query(query).Exec(); err != nil {
		panic(err)
	}
}

func (db *Cassandra) SaveUpdate(update models.Update) error {
	query := `INSERT INTO wiki_users (id, user, bot) VALUES (uuid(), ?, ?) IF NOT EXISTS`
	if err := db.Session.Query(query, update.User, update.Bot).Exec(); err != nil {
		fmt.Println("Error inserting wiki_user:", err)
		return err
	}
	query = `SELECT id FROM wiki_users WHERE user = ? AND bot = ? LIMIT 1`
	var userID gocql.UUID
	if err := db.Session.Query(query, update.User, update.Bot).Scan(&userID); err != nil {
		fmt.Println("Error querying wiki_users:", err)
		return err
	}
	query = `INSERT INTO uris (id, uri) VALUES (uuid(), ?) IF NOT EXISTS`
	if err := db.Session.Query(query, update.Uri).Exec(); err != nil {
		fmt.Println("Error inserting URI:", err)
		return err
	}
	var uriID gocql.UUID
	query = `SELECT id FROM uris WHERE uri = ? LIMIT 1`
	if err := db.Session.Query(query, update.Uri).Scan(&uriID); err != nil {
		fmt.Println("Error querying URIs:", err)
		return err
	}
	query = `INSERT INTO updates (id, uri_id, user_id) VALUES (uuid(), ?, ?)`
	if err := db.Session.Query(query, uriID, userID).Exec(); err != nil {
		fmt.Println("Error inserting update:", err)
		return err
	}
	return nil
}

func (db *Cassandra) GetStatistics() (*models.Statistics, error) {
	stats := db.stats
	query := `SELECT COUNT(*) FROM updates`
	if err := db.Session.Query(query).Scan(&stats.Messages); err != nil {
		fmt.Println("Error querying updates:", err)
	}
	query = `SELECT COUNT(*) FROM uris`
	if err := db.Session.Query(query).Scan(&stats.Urls); err != nil {
		fmt.Println("Error querying statistics:", err)
	}
	query = `SELECT COUNT(*) FROM wiki_users WHERE bot = true`
	if err := db.Session.Query(query).Scan(&stats.Bots); err != nil {
		fmt.Println("Error querying bot statistics:", err)
	}
	query = `SELECT COUNT(*) FROM wiki_users WHERE bot = false`
	if err := db.Session.Query(query).Scan(&stats.NonBots); err != nil {
		fmt.Println("Error querying non-bot statistics:", err)
	}
	return &stats, nil
}
