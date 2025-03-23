package token

import "time"

type Maker interface {
	CreateToken(userID int, duration time.Duration) (string, error)

	VerifyToken(token string) (*Payload, error)
}
