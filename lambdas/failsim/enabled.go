//go:build ENABLE_FAILSIM
// +build ENABLE_FAILSIM

package failsim

import (
	"errors"
	"math/rand/v2"
)

func OopsFailed() error {
	if rand.IntN(100) == 60 {
		return errors.New("fail simulation")
	}

	return nil
}
