package web

import (
	"sort"
	"testing"
)

func TestFindStaticAssets(t *testing.T) {
	res, err := findStaticAssets("web/static/test", "test")
	if err != nil {
		t.Fatal(err)
	}
	ss := sort.StringSlice(res)
	ss.Sort()
	res = []string(ss)
	expected := []string{"test/somedir/a.test", "test/stuff.test"}
	if len(expected) != len(res) {
		t.Fatalf("Result differs from expected: %s vs %s", res, expected)
	}
	for i, s := range expected {
		if res[i] != s {
			t.Errorf("Pos %d differs from expected: %s vs %s", i, res[i], s)
		}
	}
}