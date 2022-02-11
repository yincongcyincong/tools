package utils

func GetBoolMap(strs []string) map[string]bool {
	res := make(map[string]bool)
	for _, s := range strs {
		res[s] = true
	}

	return res
}
