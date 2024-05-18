//go:build !ENABLE_FAILSIM
// +build !ENABLE_FAILSIM

package failsim

func OopsFailed() error {
	return nil
}
