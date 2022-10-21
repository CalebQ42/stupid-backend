package stupid

import (
	"time"

	"github.com/pascaldekloe/jwt"
)

func (b Backend) verifyToken(token string) (uuid string) {
	claim, err := jwt.EdDSACheck([]byte(token), b.pubKey)
	if err != nil {
		return
	}
	if claim.Expires.Time().After(time.Now()) {
		return
	}
	if !claim.NotBefore.Time().After(time.Now()) {
		return
	}
	uuid = claim.Subject
	return
}

func (b Backend) createToken(uuid string) (token []byte, err error) {
	var claim jwt.Claims
	claim.Subject = uuid
	claim.Issued = jwt.NewNumericTime(time.Now().Round(time.Second))
	return claim.EdDSASign(b.privKey)
}
