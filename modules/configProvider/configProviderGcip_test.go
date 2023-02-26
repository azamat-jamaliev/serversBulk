package configProvider

import "testing"

func TestGetGetAnsibleSshCreds(t *testing.T) {

	srvsCreds, err := getGetAnsibleSshCreds()
	if err != nil || len(srvsCreds) < 1 {
		t.Fatalf(`TestGetGetAnsibleSshCreds Failed, error=%v`, err)
	}

}
