package utils

import "os"

// checks is file exists
func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
