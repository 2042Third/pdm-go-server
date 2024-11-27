package util

import "strconv"

func ToInt(s string) (int, error) {
	return strconv.Atoi(s)
}
