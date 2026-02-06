// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package randomtest

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

var _ knownvalue.Check = stringLengthExact{}

type stringLengthExact struct {
	length int
}

// CheckValue determines whether the passed value is of type string, and
// contains a matching length of bytes.
func (v stringLengthExact) CheckValue(other any) error {
	otherVal, ok := other.(string)

	if !ok {
		return fmt.Errorf("expected string value for StringLengthExact check, got: %T", other)
	}

	if len(otherVal) != v.length {
		return fmt.Errorf("expected string of length %d for StringLengthExact check, got: %d (value = %s)", v.length, len(otherVal), otherVal)
	}

	return nil
}

// String returns the string representation of the value.
func (v stringLengthExact) String() string {
	return strconv.FormatInt(int64(v.length), 10)
}

// StringLengthExact returns a Check for asserting the exact length of the
// value passed to the CheckValue method.
func StringLengthExact(length int) stringLengthExact {
	return stringLengthExact{
		length: length,
	}
}
