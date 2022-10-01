package stupid

type CreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type CreateReturn struct {
	ID      string `json:"_id"`
	Token   string `json:"token"`
	Problem string `json:"problem"`
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthReturn struct {
	ID      string `json:"_id"`
	Token   string `json:"token"`
	Timeout int    `json:"timeout"`
}
