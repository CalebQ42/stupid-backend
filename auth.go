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
	Permissions    map[string]bool `json:"permissions" bson:"permissions"`
	ID             string          `json:"id" bson:"_id"`
	AppID          string          `json:"appID" bson:"appID"`
	Alias          string          `json:"alias" bson:"alias"`
	AllowedDomains []string        `json:"allowedDomains" bson:"allowedDomains"`
	Death          int64           `json:"death" bson:"death"`
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

func (s *Stupid) handleToken(r *Request) (err error) {
	t, ok := r.Query["token"]
	if !ok {
		return
	}
	if len(t) != 1 {
		return
	}
	r.User, err = s.decodeJWT(t[0])
	return
}

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
	out, err := json.Marshal(map[string]string{
		"token":   token,
		"problem": "",
	})
	if err != nil {
		fmt.Printf("error while marshalling response: %s", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Resp.WriteHeader(http.StatusCreated)
	_, err = r.Resp.Write(out)
	if err != nil {
		fmt.Printf("error while writing response: %s", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func writeCreateUserProblemResp(problem string, w http.ResponseWriter) {
	resp := map[string]string{
		"token":   "",
		"problem": problem,
	}
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
		timeoutTime := 3 ^ ((authUsr.Failed / 3) - 1)
		if timeoutTime > 60 {
			timeoutTime = 60
		}
		tim := time.Unix(authUsr.LastTimeout, 0)
		tim = tim.Add(time.Duration(timeoutTime) * time.Minute)
		if tim.After(time.Now()) {
			remain := time.Until(tim).Round(time.Minute)
			var out []byte
			out, err = json.Marshal(map[string]any{
				"token":   "",
				"timeout": int64(remain.Minutes()),
			})
			if err != nil {
				fmt.Printf("error while marshalling auth response: %s", err)
				r.Resp.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = r.Resp.Write(out)
			if err != nil {
				fmt.Printf("error while writing auth response: %s", err)
				r.Resp.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}
	pass, err = hashPassword(pass, authUsr.Salt)
	if err != nil {
		fmt.Printf("error while hashing password: %s", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	if pass != authUsr.Password {
		timeoutTime := -1
		if authUsr.Failed+1%3 == 0 {
			err = s.users.IncrementAndUpdateLastTimeout(authUsr.ID, time.Now().Unix())
			timeoutTime = 3 ^ ((authUsr.Failed / 3) - 1)
			if timeoutTime > 60 {
				timeoutTime = 60
			}
		} else {
			err = s.users.IncrementFailed(authUsr.ID)
		}
		if err != nil {
			fmt.Printf("error while incrementing failed: %s", err)
			r.Resp.WriteHeader(http.StatusInternalServerError)
			return
		}
		var out []byte
		out, err = json.Marshal(map[string]any{
			"token":   "",
			"timeout": timeoutTime,
		})
		if err != nil {
			fmt.Printf("error while marshalling auth response: %s", err)
			r.Resp.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = r.Resp.Write(out)
		if err != nil {
			fmt.Printf("error while writing auth response: %s", err)
			r.Resp.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	token, err := s.generateJWT(authUsr)
	if err != nil {
		fmt.Printf("error while generating jwt: %s", err)
		r.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	out, err := json.Marshal(map[string]any{
		"token":   token,
		"timeout": -1,
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
	}
}
