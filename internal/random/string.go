// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package random

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sort"
)

type StringParams struct {
	Length          int64
	Upper           bool
	MinUpper        int64
	Lower           bool
	MinLower        int64
	Numeric         bool
	MinNumeric      int64
	Special         bool
	MinSpecial      int64
	OverrideSpecial string
}

func CreateString(input StringParams) ([]byte, error) {
	const numChars = "0123456789"
	const lowerChars = "abcdefghijklmnopqrstuvwxyz"
	const upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialChars := "!@#$%&*()-_=+[]{}<>:?"
	var result []rune

	if input.OverrideSpecial != "" {
		specialChars = input.OverrideSpecial
	}

	chars := ""
	if input.Upper {
		chars += upperChars
	}
	if input.Lower {
		chars += lowerChars
	}
	if input.Numeric {
		chars += numChars
	}
	if input.Special {
		chars += specialChars
	}

	if chars == "" {
		return nil, errors.New("the character set specified is empty")
	}

	minMapping := map[string]int64{
		numChars:     input.MinNumeric,
		lowerChars:   input.MinLower,
		upperChars:   input.MinUpper,
		specialChars: input.MinSpecial,
	}

	result = make([]rune, 0, input.Length)

	for k, v := range minMapping {
		s, err := generateRandomBytes(&k, v)
		if err != nil {
			return nil, err
		}
		result = append(result, s...)
	}

	s, err := generateRandomBytes(&chars, input.Length-int64(len(result)))
	if err != nil {
		return nil, err
	}

	result = append(result, s...)

	order := make([]byte, len(result))
	if _, err := rand.Read(order); err != nil {
		return nil, err
	}

	sort.Slice(result, func(i, j int) bool {
		return order[i] < order[j]
	})

	return []byte(string(result)), nil
}

func generateRandomBytes(charSet *string, length int64) ([]rune, error) {
	if charSet == nil {
		return nil, errors.New("charSet is nil")
	}

	if *charSet == "" && length > 0 {
		return nil, errors.New("charSet is empty")
	}

	runeSet := []rune(*charSet)

	bytes := make([]rune, length)
	setLen := big.NewInt(int64(len(runeSet)))
	for i := range bytes {
		idx, err := rand.Int(rand.Reader, setLen)
		if err != nil {
			return nil, err
		}
		character := runeSet[idx.Int64()]
		bytes[i] = character
	}
	return bytes, nil
}
