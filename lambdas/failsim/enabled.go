//go:build ENABLE_FAILSIM
// +build ENABLE_FAILSIM

package failsim

import (
	"errors"
	"math/rand"
)

// function signature with this body will be built
// if ENABLE_FAILSIM is defined via the Go default compiler

// Invoking the failure simulator will result
// in erroring if the PRNG generates exactly 60
// (could have been whatever number: uniform dist.)
// otherwise, success (enabled: do not change
// the first two lines of comments to allow Go
// compiler to decide which code to compile from
// passed compile flags)
func OopsFailed() error {
	if rand.Intn(100) == 60 {
		return errors.New("fail simulation")
	}

	return nil
}
