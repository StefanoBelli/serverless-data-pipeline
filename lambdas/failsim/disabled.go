//go:build !ENABLE_FAILSIM
// +build !ENABLE_FAILSIM

package failsim

// function signature with this body will be built
// if ENABLE_FAILSIM is NOT defined via the Go default compiler

// Invoking the failure simulator will result
// in always success (disabled: do not change
// the first two lines of comments to allow Go
// compiler to decide which code to compile from
// passed compile flags)
func OopsFailed() error {
	return nil
}
