package stupid

type ApiKey struct {
	Features map[string]bool `bson:"features"`
	Key      string          `bson:"_id"`
	AppID    string          `bson:"appID"`
	Alias    string          `bson:"alias"`
	Death    int             `bson:"death"`
}
