package logging

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap/zapcore"
	"time"
)

type MongoCore struct {
	collection *mongo.Collection
	level      zapcore.Level
}

func NewMongoCore(uri, db, coll string, lvl zapcore.Level) (*MongoCore, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return &MongoCore{
		collection: client.Database(db).Collection(coll),
		level:      lvl,
	}, nil
}

func (m *MongoCore) Enabled(lvl zapcore.Level) bool {
	return lvl >= m.level
}

func (m *MongoCore) With(fields []zapcore.Field) zapcore.Core {
	return m
}

func (m *MongoCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if m.Enabled(ent.Level) {
		return ce.AddCore(ent, m)
	}
	return ce
}

func (m *MongoCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	doc := map[string]interface{}{
		"level":     ent.Level.String(),
		"message":   ent.Message,
		"timestamp": ent.Time.UTC(),
	}

	encoder := zapcore.NewMapObjectEncoder()
	for _, f := range fields {
		f.AddTo(encoder)
	}

	for k, v := range encoder.Fields {
		doc[k] = v
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.collection.InsertOne(ctx, doc)
	return err
}

func (m *MongoCore) Sync() error {
	return nil
}
