//go:build production

package assets

import "errors"

func EsbuildAvailable() bool {
	return false
}

func BuildAssets(_ BuildOptions) error {
	return errors.New("can't build browser assets in production mode")
}
