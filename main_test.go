package main

import (
	"testing"
)

func TestStringSizeToBytes(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		wants int64
		err   error
	}{
		{
			name:  "1G",
			input: "1G",
			wants: 1e9,
			err:   nil,
		},
		{
			name:  "1M",
			input: "1M",
			wants: 1e6,
			err:   nil,
		},
		{
			name:  "1K",
			input: "1K",
			wants: 1e3,
			err:   nil,
		},
		{
			name:  "1",
			input: "1",
			wants: int64(1),
			err:   nil,
		},
		{
			name:  "1g",
			input: "1g",
			wants: 1e9,
			err:   nil,
		},
		{
			name:  "1m",
			input: "1m",
			wants: 1e6,
			err:   nil,
		},
		{
			name:  "1k",
			input: "1k",
			wants: 1e3,
			err:   nil,
		},
		{
			name:  "1P",
			input: "1P",
			wants: int64(-1),
			err:   ErrImproperByteStringDenomination,
		},
		{
			name:  "GG",
			input: "GG",
			wants: int64(-1),
			err:   ErrInteger64Conversion,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := convertStringSizeToBytes(tc.input)

			if res != tc.wants {
				t.Logf("convert string size results wanted: %v got: %v", tc.wants, res)
				t.Fail()
			} else if err != tc.err {
				t.Logf("convert string to size results wanted: %v got: %v", tc.err, err)
				t.Fail()
			}
		})
	}

}
