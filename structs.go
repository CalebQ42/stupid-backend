package stupid

type ApiKey struct {
	ID       string `json:"_id"`
	Features string `json:"features"`
	Death    int64  `json:"death"`
}

type GlobalUser struct {
	ID          string `json:"_id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	Failed      int    `json:"failed"`
	LastTimeout int64  `json:"lastTimout"`
}

type AppUser struct {
	ID            string `json:"_id"`
	Global        bool   `json:"hasGlobal"`
	LastConnected int    `json:"lastConnected"`
}

type AppData struct {
	Data any    `json:"data"`
	ID   string `json:"_id"`
	Name string `json:"displayName"`
	Type string `json:"type"`
}

type UserData struct {
	Data       any
	ID         string   `json:"_id"`
	Owner      string   `json:"owner"`
	Read       []string `json:"readPerm"`
	Write      []string `json:"writePerm"`
	GlobalRead bool     `json:"globalRead"`
}

type Report struct {
	ID     string `json:"_id"`
	Stack  string `json:"stack"`
	Action string `json:"action"`
}

type Crash struct {
	ID      string   `json:"_id"`
	Reports []Report `json:"reports"`
}
