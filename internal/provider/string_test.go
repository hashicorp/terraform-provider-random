package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResourcePasswordStringStateUpgradeV1(t *testing.T) {
	cases := []struct {
		name            string
		stateV1         map[string]interface{}
		err             error
		expectedStateV2 map[string]interface{}
	}{
		{
			name:    "raw state is nil",
			stateV1: nil,
			err:     errors.New("state upgrade failed, state is nil"),
		},
		{
			name:    "number is not bool",
			stateV1: map[string]interface{}{"number": 0},
			err:     errors.New("state upgrade failed, number is not a boolean: int"),
		},
		{
			name:            "success",
			stateV1:         map[string]interface{}{"number": true},
			expectedStateV2: map[string]interface{}{"number": true, "numeric": true},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actualStateV2, err := resourcePasswordStringStateUpgradeV1(context.Background(), c.stateV1, nil)

			if c.err != nil {
				if !cmp.Equal(c.err.Error(), err.Error()) {
					t.Errorf("expected: %q, got: %q", c.err.Error(), err)
				}
				if !cmp.Equal(c.expectedStateV2, actualStateV2) {
					t.Errorf("expected: %+v, got: %+v", c.expectedStateV2, err)
				}
			} else {
				if err != nil {
					t.Errorf("err should be nil, actual: %v", err)
				}

				if !cmp.Equal(actualStateV2, c.expectedStateV2) {
					t.Errorf("expected: %v, got: %v", c.expectedStateV2, actualStateV2)
				}
			}
		})
	}
}
