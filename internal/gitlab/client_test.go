package gitlab

import "testing"

func TestMapStatusErr(t *testing.T) {
	cases := []int{401, 403, 404, 429, 400, 500}
	for _, c := range cases {
		err := mapStatusErr(c, "x")
		if err == nil {
			t.Fatalf("expected error for %d", c)
		}
	}
}

func TestSanitizePagination(t *testing.T) {
	page, perPage := sanitizePagination(-1, 999)
	if page != 1 || perPage != 100 {
		t.Fatalf("unexpected pagination: %d %d", page, perPage)
	}
}
