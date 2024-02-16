package util

import (
	"os"
)

func EnsureFolderExists(p string) error {
	_, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(p, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
