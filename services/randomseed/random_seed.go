package randomseed

import "strconv"

func RandomSeedToBytes(randomSeed uint64) []byte {
	return []byte(strconv.FormatUint(randomSeed, 10))
}
