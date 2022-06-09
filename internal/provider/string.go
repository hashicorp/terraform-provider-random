package provider

import (
	"crypto/rand"
	"math/big"
	"sort"
)

type randomStringParams struct {
	length          int64
	upper           bool
	minUpper        int64
	lower           bool
	minLower        int64
	numeric         bool
	minNumeric      int64
	special         bool
	minSpecial      int64
	overrideSpecial string
}

func createRandomString(input randomStringParams) ([]byte, error) {
	const numChars = "0123456789"
	const lowerChars = "abcdefghijklmnopqrstuvwxyz"
	const upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var specialChars = "!@#$%&*()-_=+[]{}<>:?"
	var result []byte

	if input.overrideSpecial != "" {
		specialChars = input.overrideSpecial
	}

	var chars = ""
	if input.upper {
		chars += upperChars
	}
	if input.lower {
		chars += lowerChars
	}
	if input.numeric {
		chars += numChars
	}
	if input.special {
		chars += specialChars
	}

	minMapping := map[string]int64{
		numChars:     input.minNumeric,
		lowerChars:   input.minLower,
		upperChars:   input.minUpper,
		specialChars: input.minSpecial,
	}

	result = make([]byte, 0, input.length)

	for k, v := range minMapping {
		s, err := generateRandomBytes(&k, v)
		if err != nil {
			return nil, err
		}
		result = append(result, s...)
	}

	s, err := generateRandomBytes(&chars, input.length-int64(len(result)))
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

	return result, nil
}

func generateRandomBytes(charSet *string, length int64) ([]byte, error) {
	bytes := make([]byte, length)
	setLen := big.NewInt(int64(len(*charSet)))
	for i := range bytes {
		idx, err := rand.Int(rand.Reader, setLen)
		if err != nil {
			return nil, err
		}
		bytes[i] = (*charSet)[idx.Int64()]
	}
	return bytes, nil
}
