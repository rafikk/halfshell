package util

func FirstString(str ...string) (s string) {
	for _, s := range str {
		if s != "" {
			return s
		}
	}
	return s
}

func FirstUInt(ints ...uint64) (n uint64) {
	for _, n := range ints {
		if n > 0 {
			return n
		}
	}
	return n
}
