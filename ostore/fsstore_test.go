package ostore

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFSStoreGetPutDelete(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "fsstoragetest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	st := &FSStore{}
	if err := st.Initialize(map[string]string{FS_ROOT_PATH: tmpDir}); err != nil {
		t.Fatal("Unexpected error: %s", err)
	}
	runTestStoreGetPutDelete("FSStore", st, t)
}
