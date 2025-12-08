package random

import (
	"fmt"
	"math/rand"
	"strings"
)

var (
	letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	digits  = []byte("0123456789")
)

func String(lengthOpt ...int) string {
	length := 10
	if len(lengthOpt) == 1 {
		length = lengthOpt[0]
	}
	alphanum := letters
	alphanum = append(alphanum, digits...)
	s := make([]byte, length)
	for i := 0; i < length; i++ {
		s[i] = alphanum[rand.Intn(len(alphanum))]
	}
	return string(s)
}

func Email() string {
	return strings.ToLower(fmt.Sprintf("%s@%s.%s", Letters(6), Letters(6), Letters(3)))
}

func Letters(length int) string {
	l := ""
	for i := 0; i < length; i++ {
		l += Letter()
	}
	return l
}

func Digits(length int) string {
	d := ""
	for i := 0; i < length; i++ {
		d += Digit()
	}
	return d
}

func Digit() string {
	d := digits[rand.Intn(len(digits))]
	return string(d)
}

func Letter() string {
	l := letters[rand.Intn(len(letters))]
	return string(l)
}

func IntSlice(lenght int) []int {
	var s []int
	for i := 0; i < lenght; i++ {
		s = append(s, rand.Intn(10))
	}
	return s
}

func Int(min, max int) int {
	if max < min {
		min, max = max, min
	}

	return rand.Intn(max-min+1) + min
}
