package util

func FirstString(str ...string) (s string) {
	for _, s := range str {
		if s != "" {
			return s
		}
	}
	return s
}
