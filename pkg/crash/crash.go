package crash

type Individual struct {
	ID       string `json:"id" bson:"_id"`
	Error    string `json:"error" bson:"error"`
	Platform string `json:"platform" bson:"platform"`
	Stack    string `json:"stack" bson:"stack"`
}

type Group struct {
	ID        string       `json:"id" bson:"_id"`
	Error     string       `json:"error" bson:"error"`
	FirstLine string       `json:"first" bson:"first"`
	Crashes   []Individual `json:"crashes" bson:"crashes"`
}
