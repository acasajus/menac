package ostore

import (
	"bytes"
	"testing"
	"time"
)

const (
	TEST_ACCESS_KEY   = "DUV2M1HTERVY22GD8IDL"
	TEST_SECRET_KEY   = "jIDyWkZe+CYZfSlzaPNix4rgapMiRSIQzuxK5MKI"
	TEST_ENDPOINT_URL = "http://grub85.ecm.ub.es"
)

func getS3InitParams() map[string]string {
	return map[string]string{
		S3_ACCES_KEY:    TEST_ACCESS_KEY,
		S3_SECRET_KEY:   TEST_SECRET_KEY,
		S3_ENDPOINT_URL: TEST_ENDPOINT_URL,
		S3_BUCKET:       "menac_s3_store",
	}
}

func TestS3StoreGetPutDelete(t *testing.T) {
	st := &S3Store{}
	if err := st.Initialize(getS3InitParams()); err != nil {
		t.Fatal("Unexpected error: %s", err)
	}
	runTestStoreGetPutDelete("S3Store", st, t)
}

func TestS3StoreInitialize(t *testing.T) {
	initParams := getS3InitParams()
	st := &S3Store{}
	//Make sure bucket is there
	if err := st.Initialize(initParams); err != nil {
		t.Fatal("Unexpected error: %s", err)
	}
	//Delete
	if err := st.bucket.DelBucket(); err != nil {
		t.Errorf("Could not delete bucket: %s", err)
	}
	//Create bucket
	if err := st.Initialize(initParams); err != nil {
		t.Fatal("Unexpected error: %s", err)
	}
	//Init if bucket is there
	if err := st.Initialize(initParams); err != nil {
		t.Fatal("Unexpected error: %s", err)
	}
}

func runTestStoreGetPutDelete(name string, st Store, t *testing.T) {
	so := &StoredObject{
		Hash:       "INVALID",
		Type:       "test",
		Expiration: time.Now().Add(time.Hour),
	}
	//MAke sure object is not in the store
	if err := st.Delete(so); err != nil && err != ErrNotExists {
		t.Fatalf("[%s] Cannot delete data from store: %s", name, err)
	}
	byteData := []byte(HASH_TEST_DATA)
	if err := st.Put(so, bytes.NewBufferString(HASH_TEST_DATA), int64(len(byteData))); err != ErrHashMismatch {
		if err == nil {
			t.Fatalf("[%s] Could write data with incorrect hash!", name)
		} else {
			t.Fatalf("[%s] Cannot write to store: %s", name, err)
		}
	}
	out := &bytes.Buffer{}
	if err := st.Get(so, out); err != ErrNotExists {
		t.Fatalf("[%s] Cannot get data from store: %s", name, err)
	}
	so.Hash = HASH_EXPECTED
	if err := st.Put(so, bytes.NewBufferString(HASH_TEST_DATA), int64(len(byteData))); err != nil {
		t.Fatalf("[%s] Cannot write to store: %s", name, err)
	}
	if err := st.Get(so, out); err != nil {
		t.Fatalf("[%s] Cannot get data from store: %s", name, err)
	}
	if !bytes.Equal(out.Bytes(), byteData) {
		t.Errorf("[%s] Out data differs from in data", name)
	}
	if err := st.Delete(so); err != nil {
		t.Fatalf("[%s] Cannot delete data from store: %s", name, err)
	}
}
