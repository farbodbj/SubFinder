//go:build !linux
// +build !linux

package dialer

func setReusePort(fd uintptr) error {
	return nil
}
