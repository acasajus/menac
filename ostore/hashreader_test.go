package ostore

import (
	"bytes"
	"testing"
)

const (
	HASH_TEST_DATA = `-----BEGIN CERTIFICATE-----
MIIEozCCA4ugAwIBAgICH3AwDQYJKoZIhvcNAQELBQAwQzESMBAGCgmSJomT8ixk
ARkWAmVzMRgwFgYKCZImiZPyLGQBGRYIaXJpc2dyaWQxEzARBgNVBAMTCklSSVNH
cmlkQ0EwHhcNMTQwNjAyMTQ1MTMxWhcNMTUwNjAyMTQ1MTMxWjBYMRIwEAYKCZIm
iZPyLGQBGRYCZXMxGDAWBgoJkiaJk/IsZAEZFghpcmlzZ3JpZDEPMA0GA1UEChMG
ZWNtLXViMRcwFQYDVQQDEw5BZHJpYW4tQ2FzYWp1czCCASIwDQYJKoZIhvcNAQEB
BQADggEPADCCAQoCggEBAL51Qe0N28ZhmQetQE8LcNN84l20UndrZtyvzac5ohgv
1kZ7pzwgX+S6w71LxT3wXcDzI21T3IJVkDhnbzBspWjY7vgps/qYRcPYJqCjVi+s
HEC6Q0ZlmtdPGUdaYsp+zCUuH6ZAIcv/q0BvcUdXN6tpLR4Su1FYdf0cJJxM3QMa
Lz1AWATV9YHQ/V/FS/0EKob5fBEyiHynvXBhCjsJE/0/uAU1NGT1K8+VEJ74PtlK
wTzmydgmbowaI6j7P6K2I97DGBx1S+g8m2ebB3hNoGQHjRmZyYPy/Z74/5kJLQ3a
py8CjcVAJrzFFW242GVUFBLLvfkvnDsowmfirxMBxGUCAwEAAaOCAYowggGGMAwG
A1UdEwEB/wQCMAAwHQYDVR0OBBYEFK/QwtnkS1TYAdnNBfvvzsN85PCLMGsGA1Ud
IwRkMGKAFJ1CZC5bqV7x/2zIPDbT8bZDmK0aoUekRTBDMRIwEAYKCZImiZPyLGQB
GRYCZXMxGDAWBgoJkiaJk/IsZAEZFghpcmlzZ3JpZDETMBEGA1UEAxMKSVJJU0dy
aWRDQYIBADAOBgNVHQ8BAf8EBAMCBLAwHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsG
AQUFBwMEMDMGA1UdEQQsMCqGKGh0dHA6Ly9wa2kuaXJpc2dyaWQuZXMvYXVkaXQv
P2lkPWExOWIxYzkwIgYDVR0SBBswGYYXaHR0cDovL3BraS5pcmlzZ3JpZC5lcy8w
OAYDVR0fBDEwLzAtoCugKYYnaHR0cDovL3BraS5pcmlzZ3JpZC5lcy9jYS9jcmwv
Y2FjcmwuY3JsMCgGA1UdIAQhMB8wDwYNKwYBBAG6ewICBAEEADAMBgoqhkiG90wF
AgIBMA0GCSqGSIb3DQEBCwUAA4IBAQASf16sYp0Hiyuf+0s/B7u1HEFqKnySs4MY
lUbt1AkiLUOoS6gPVOw8Nan0Elebz2MjLDvGse2jdU83JwPYl+Zi5iDcBHz7va7l
kvlszc4SDtYqwUngJltOyrP9czchbLpMivqqOVEeFHEXXVUFLS3em6rYOjGbGRxR
Ih/EeBJC71HEM8RH46QI01IOxVNSiioG5vnDfCHoZLRDNsDLoh9GTuEGGo+Wf+S8
LmJbi/XocWalO+aJfoOUf/F4p7jy2gqoWgHx2ausiAOcYklLW3OfSXEj4fiS0gpf
/FD6AYjn5OeFBPuv/HC28equU/pdT6UUYWOho5eol8iExQlMx709
-----END CERTIFICATE-----`
	HASH_EXPECTED = "dfc509c952649dc27e9e38f8d226bd07e7027216edd95a60c731d8f7d306cdc60bd4626edd6477db20a464878a9ad0a99d8ed9568198df519d28ff8b154028ad"
)

func TestHashReader(t *testing.T) {
	data := []byte(HASH_TEST_DATA)
	hr := NewHashReader(bytes.NewBufferString(HASH_TEST_DATA))
	if n, err := hr.Read(data); err != nil {
		t.Error("Could not read from HashReader: %s", err)
	} else if n != len(data) {
		t.Errorf("Could not read all required data from hash reader: %d vs expected %d", n, len(data))
	}
	if hr.HexDigest() != HASH_EXPECTED {
		t.Error("Unexpected hash digest result: %s vs expected %s", hr.HexDigest(), HASH_EXPECTED)
	}
}
