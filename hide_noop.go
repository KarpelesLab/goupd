//go:build !windows
// +build !windows

package goupd

func hideFile(path string) error {
	return nil
}
