package stupid

type ApiKey struct {
	Features string `json:"features"`
	Death    int64  `json:"death"`
}

type GlobalUser struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	Failed      int    `json:"failed"`
	LastTimeout int64  `json:"lastTimout"`
}

type AppUser struct {
	Global        bool `json:"hasGlobal"`
	LastConnected int  `json:"lastConnected"`
}

type AppData struct {
	Data any    `json:"data"`
	Name string `json:"displayName"`
	Type string `json:"type"`
}

type UserData struct {
	Data       any
	Owner      string   `json:"owner"`
	Read       []string `json:"readPerm"`
	Write      []string `json:"writePerm"`
	GlobalRead bool     `json:"globalRead"`
}

type Report struct {
	Stack  string `json:"stack"`
	Action string `json:"action"`
}

type Crash struct {
	Reports []Report `json:"reports"`
}
