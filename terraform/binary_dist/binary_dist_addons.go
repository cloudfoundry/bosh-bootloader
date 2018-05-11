package binary_dist

import "os"

// the way we use this, errors should not happen
func MustAssetInfo(name string) os.FileInfo {
	info, err := AssetInfo(name)
	if err != nil {
		panic(err)
	}
	return info
}
