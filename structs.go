package stupid

type ApiKey struct {
	ID       string
	Features string
	Death    int64
}

type GlobalUser struct {
	ID          string
	Username    string
	Password    string
	Email       string
	Failed      int
	LastTimeout int64
}

type AppUser struct {
	ID            string
	Global        bool
	LastConnected int
}

type AppData struct {
	ID   string
	Name string
	Type string
	Data any
}

type UserData struct {
	ID         string
	Owner      string
	GlobalRead bool
	Read       []string
	Write      []string
	Data       any
}

type Report struct {
	ID     string
	Stack  string
	Action string
}

type Crash struct {
	ID      string
	Reports []Report
}
