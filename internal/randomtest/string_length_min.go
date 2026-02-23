// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package randomtest

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

var _ knownvalue.Check = stringLengthMin{}

type stringLengthMin struct {
	minLength int
}

// CheckValue determines whether the passed value is of type string, and
// contains a matching length of characters (runes).
func (v stringLengthMin) CheckValue(other any) error {
	otherVal, ok := other.(string)

	if !ok {
		return fmt.Errorf("expected string value for StringLengthMin check, got: %T", other)
	}

	runeCount := utf8.RuneCountInString(otherVal)
	if runeCount < v.minLength {
		return fmt.Errorf("expected string length to be at least %d for StringLengthMin check, got: %d (value = %s)", v.minLength, runeCount, otherVal)
	}

	return nil
}

// String returns the string representation of the value.
func (v stringLengthMin) String() string {
	return strconv.FormatInt(int64(v.minLength), 10)
}

// StringLengthMin returns a Check for asserting the minimum length of the
// value passed to the CheckValue method.
func StringLengthMin(minLength int) stringLengthMin {
	return stringLengthMin{
		minLength: minLength,
	}
}
