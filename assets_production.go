//go:build production

package guiapi

import "errors"

func EsbuildAvailable() bool {
	return false
}

func BuildAssets() error {
	return errors.New("can't build browser assets in production mode")
}
