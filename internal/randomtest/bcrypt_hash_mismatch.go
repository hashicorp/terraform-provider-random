// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package randomtest

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"golang.org/x/crypto/bcrypt"
)

var _ compare.ValueComparer = bcryptHashMismatch{}

type bcryptHashMismatch struct{}

// CompareValues determines whether the first value is a valid bcrypt hash of the second value.
func (v bcryptHashMismatch) CompareValues(values ...any) error {
	if len(values) != 2 {
		return fmt.Errorf("expected to receive two values to compare, but got: %d", len(values))
	}

	hash, ok := values[0].(string)
	if !ok {
		return fmt.Errorf("expected bcrypt hash to be of type string, but got: %T", values[0])
	}

	plainTextVal, ok := values[1].(string)
	if !ok {
		return fmt.Errorf("expected plain text value to be of type string, but got: %T", values[1])
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainTextVal))
	if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return fmt.Errorf("unexpected error: %s", err)
	}

	return nil
}

// BcryptHashMismatch returns a ValueComparer for asserting that the first value in the sequence is not a matching
// bcrypt hash of the second value. If there is an error parsing the hash, this compare will fail.
func BcryptHashMismatch() bcryptHashMismatch {
	return bcryptHashMismatch{}
}
