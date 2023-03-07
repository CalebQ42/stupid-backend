package stupid

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/CalebQ42/stupid-backend/pkg/db"
	"github.com/google/uuid"
	"github.com/pascaldekloe/jwt"
	"golang.org/x/crypto/argon2"
)

type apiKey struct {
	Permissions map[string]bool `json:"permissions" bson:"permissions"`
	ID          string          `json:"id" bson:"_id"`
	AppID       string          `json:"appID" bson:"appID"`
	Alias       string          `json:"alias" bson:"alias"`
	Death       int64           `json:"death" bson:"death"`
}

func hashPassword(password, salt string) (string, error) {
	saltByts, err := base64.RawStdEncoding.DecodeString(salt)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(password), saltByts, 1, 64*1024, 4, 32)), nil
}

func generateSalt() (string, error) {
	hold := make([]byte, 16)
	_, err := rand.Read(hold)
	return base64.RawStdEncoding.EncodeToString(hold), err
}

func (s *Stupid) generateJWT(u *user) (string, error) {
	var c jwt.Claims
	c.Subject = u.ID
	c.Issued = jwt.NewNumericTime(time.Now().Round(time.Second))
	c.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	out, err := c.EdDSASign(s.userPriv)
	fmt.Println("out token", string(out))
	return string(out), err
}

var ErrExpired = errors.New("token is expired")

// Returns jwt.ErrSigMiss on bad token, db.ErrNotFound if id is invalid, and ErrExpired if expired.
func (s *Stupid) decodeJWT(token string) (*AuthdUser, error) {
	c, err := jwt.EdDSACheck([]byte(token), s.userPub)
	if err != nil {
		return nil, err
	}
	if c.Expires.Time().Before(time.Now()) {
		return nil, ErrExpired
	}
	id := c.Subject
	usr := new(user)
	err = s.users.Get(id, usr)
	if err != nil {
		return nil, err
	}
	return &AuthdUser{
		ID:       usr.ID,
		Username: usr.Username,
		Email:    usr.Email,
	}, nil
}

type createResp struct {
	Token   string `json:"token"`
	Problem string `json:"problem"`
}

func (s *Stupid) createUser(r *Request) {
	if r.Method != http.MethodPost {
		r.Resp.WriteHeader(http.StatusBadRequest)
		return
	}
	body := make(map[string]string)
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		fmt.Printf("error while decoding create user request: %s", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	var name, pass, email string = body["username"], body["password"], body["email"]
	if name == "" || pass == "" || email == "" {
		r.Resp.WriteHeader(http.StatusBadRequest)
		return
	}
	if strings.HasPrefix(name, " ") || strings.HasSuffix(name, " ") || len(name) > 64 {
		writeCreateUserProblemResp("username", r.Resp)
		return
	}
	// TODO: Add optional bad word check to usernames.
	s.createUserMutex.Lock()
	defer s.createUserMutex.Unlock()
	usernameNotAvailable, err := s.users.Contains(map[string]any{"username": name})
	if err != nil {
		fmt.Printf("error while finding if username is already taken: %s", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	if usernameNotAvailable {
		writeCreateUserProblemResp("username", r.Resp)
		return
	}
	// TODO: check email properly
	if !strings.Contains(email, "@") && !strings.Contains(email, ".") {
		writeCreateUserProblemResp("email", r.Resp)
		return
	}
	if len(pass) < 5 || len(pass) > 32 {
		writeCreateUserProblemResp("password", r.Resp)
		return
	}
	// TODO: Check password against dictionary of common, stupid passwords (such as "Password1")
	newUser := &user{
		ID:       uuid.NewString(),
		Username: name,
		Email:    email,
	}
	newUser.Salt, err = generateSalt()
	if err != nil {
		fmt.Printf("error while generating salt: %s", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	newUser.Password, err = hashPassword(pass, newUser.Salt)
	if err != nil {
		fmt.Printf("error while generating hashing password: %s", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = s.users.Add(newUser)
	if err != nil {
		fmt.Printf("error while adding new user: %s", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	token, err := s.generateJWT(newUser)
	if err != nil {
		fmt.Printf("error while generating jwt: %s", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	out, err := json.Marshal(createResp{
		Token: token,
	})
	if err != nil {
		fmt.Printf("error while marshalling response: %s", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = r.Resp.Write(out)
	if err != nil {
		fmt.Printf("error while writing response: %s", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Resp.WriteHeader(http.StatusCreated)
}

func writeCreateUserProblemResp(problem string, w http.ResponseWriter) {
	var resp createResp
	resp.Problem = problem
	out, err := json.Marshal(resp)
	if err != nil {
		fmt.Printf("error while marshaling create user response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(out)
	if err != nil {
		fmt.Printf("error while writing response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Stupid) authUser(r *Request) {
	if r.Method != http.MethodPost {
		r.Resp.WriteHeader(http.StatusBadRequest)
		return
	}
	body := make(map[string]string)
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		fmt.Printf("error while decoding create user request: %s", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	var name, pass string = body["username"], body["password"]
	if name == "" || pass == "" {
		r.Resp.WriteHeader(http.StatusBadRequest)
		return
	}
	authUsr := new(user)
	err = s.users.Find(map[string]any{"username": name}, authUsr)
	if err == db.ErrNotFound {
		r.Resp.WriteHeader(http.StatusNotFound)
		return
	}

	if authUsr.LastTimeout != 0 {
		// TODO
	}

	pass, err = hashPassword(pass, authUsr.Salt)
	if err != nil {
		fmt.Printf("error while hashing password: %s", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	if pass != authUsr.Password {
		
	}
}
