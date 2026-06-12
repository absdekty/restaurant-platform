package hasher

import (
	"golang.org/x/crypto/bcrypt"
)

type Hasher struct {
	cost int
}

func New(cost ...int) *Hasher {
	c := 10
	if len(cost) > 0 {
		c = cost[0]
	}
	return &Hasher{cost: c}
}

func (h *Hasher) Hash(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	return string(hashed), err
}

func (h *Hasher) Compare(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
