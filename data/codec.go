package data

import (
	"wiki_updates/models"
	"wiki_updates/models/pb"

	"google.golang.org/protobuf/proto"
)

// Wire encoding for Redpanda records. Records are protobuf; the models structs
// stay as the in-process / HTTP types.

func encodeUpdate(update models.Update) ([]byte, error) {
	return proto.Marshal(&pb.Update{
		Uri:     update.Uri,
		Bot:     update.Bot,
		User:    update.User,
		MetaUri: update.Meta.Uri,
	})
}

func decodeUpdate(raw []byte) (models.Update, error) {
	var msg pb.Update
	if err := proto.Unmarshal(raw, &msg); err != nil {
		return models.Update{}, err
	}
	return models.Update{
		Uri:  msg.Uri,
		Bot:  msg.Bot,
		User: msg.User,
		Meta: models.Meta{Uri: msg.MetaUri},
	}, nil
}

func encodeUser(user models.User) ([]byte, error) {
	return proto.Marshal(&pb.User{Email: user.Email, PasswordHash: user.PasswordHash})
}

func decodeUser(raw []byte) (models.User, error) {
	var msg pb.User
	if err := proto.Unmarshal(raw, &msg); err != nil {
		return models.User{}, err
	}
	return models.User{Email: msg.Email, PasswordHash: msg.PasswordHash}, nil
}

func encodeSnapshot(snapshot statsSnapshot) ([]byte, error) {
	return proto.Marshal(&pb.StatsSnapshot{
		Messages: int64(snapshot.Messages),
		Urls:     snapshot.Urls,
		Bots:     snapshot.Bots,
		NonBots:  snapshot.NonBots,
		Offset:   snapshot.Offset,
	})
}

func decodeSnapshot(raw []byte) (statsSnapshot, error) {
	var msg pb.StatsSnapshot
	if err := proto.Unmarshal(raw, &msg); err != nil {
		return statsSnapshot{}, err
	}
	return statsSnapshot{
		Messages: int(msg.Messages),
		Urls:     msg.Urls,
		Bots:     msg.Bots,
		NonBots:  msg.NonBots,
		Offset:   msg.Offset,
	}, nil
}
