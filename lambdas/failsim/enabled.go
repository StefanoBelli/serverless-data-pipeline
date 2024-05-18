//go:build ENABLE_FAILSIM
// +build ENABLE_FAILSIM

package failsim

import (
	"errors"
	"math/rand"
)

func OopsFailed() error {
	if rand.Intn(100) == 60 {
		return errors.New("fail simulation")
	}

	return nil
}
