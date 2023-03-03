package stupid

type apiKey struct {
	Permissions map[string]bool `json:"permissions" bson:"permissions"`
	ID          string          `json:"id" bson:"_id"`
	AppID       string          `json:"appID" bson:"appID"`
	Alias       string          `json:"alias" bson:"alias"`
	Death       int64           `json:"death" bson:"death"`
}
