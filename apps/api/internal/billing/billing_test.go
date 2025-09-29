package billing

import "testing"

func TestNormalizeLimitOffset(t *testing.T) {
	cases := []struct {
		name       string
		limit      int32
		offset     int32
		wantLimit  int32
		wantOffset int32
	}{
		{name: "success: default limit when zero", limit: 0, offset: 0, wantLimit: DefaultLimit, wantOffset: 0},
		{name: "success: default limit when negative", limit: -5, offset: 10, wantLimit: DefaultLimit, wantOffset: 10},
		{name: "success: cap at max", limit: MaxLimit + 1, offset: 5, wantLimit: MaxLimit, wantOffset: 5},
		{name: "success: negative offset coerced to zero", limit: 42, offset: -1, wantLimit: 42, wantOffset: 0},
		{name: "success: exact max allowed", limit: MaxLimit, offset: 3, wantLimit: MaxLimit, wantOffset: 3},
		{name: "success: passthrough", limit: 1, offset: 0, wantLimit: 1, wantOffset: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotL, gotO := NormalizeLimitOffset(tc.limit, tc.offset)
			if gotL != tc.wantLimit || gotO != tc.wantOffset {
				t.Fatalf("NormalizeLimitOffset(%d,%d) = (%d,%d); want (%d,%d)", tc.limit, tc.offset, gotL, gotO, tc.wantLimit, tc.wantOffset)
			}
		})
	}
}
