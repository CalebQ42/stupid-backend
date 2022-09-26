package stupid

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type DefaultApp struct {
	LogColl   *mongo.Collection
	CrashColl *mongo.Collection
	appID     string
}

func NewDefaultApp(appID string, client *mongo.Client) *DefaultApp {
	return &DefaultApp{
		appID:     appID,
		LogColl:   client.Database(appID).Collection("log"),
		CrashColl: client.Database(appID).Collection("crashes"),
	}
}

func (d DefaultApp) ID() string {
	return d.appID
}

func (d DefaultApp) Logs() *mongo.Collection {
	return d.LogColl
}

func (d DefaultApp) Crashes() *mongo.Collection {
	return d.CrashColl
}

func (d *DefaultApp) Initialize() error {
	return nil
}

// TODO
type DefaultDataApp struct{}
