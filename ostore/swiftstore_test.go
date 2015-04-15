package ostore

import "testing"

const (
	TEST_SWIFT_USER      = "test:tester"
	TEST_SWIFT_PASS      = "testing"
	TEST_SWIFT_AUTH_URL  = "http://localhost:8089/auth/v1.0"
	TEST_SWIFT_CONTAINER = "swift_test_menac"
)

func getSwiftInitParams() map[string]string {
	return map[string]string{
		SWIFT_AUTH_URL:  TEST_SWIFT_AUTH_URL,
		SWIFT_USER_NAME: TEST_SWIFT_USER,
		SWIFT_API_KEY:   TEST_SWIFT_PASS,
		SWIFT_CONTAINER: TEST_SWIFT_CONTAINER,
	}
}

func TestSwiftStoreGetPutDelete(t *testing.T) {
	st := &SwiftStore{}
	if err := st.Initialize(getSwiftInitParams()); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	runTestStoreGetPutDelete("SwiftStore", st, t)
}
