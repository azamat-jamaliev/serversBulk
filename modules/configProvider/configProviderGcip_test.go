package configProvider

import (
	"fmt"
	"path"
	"testing"
)

func TestGetGetAnsibleSshCreds(t *testing.T) {

	ansibleDir := ""
	GLOBAL_GROUP_VARS_DIR := "00.global/group_vars/all"

	if ansibleDir = "../../test/ansible-inventory/CLASSIC"; exists(path.Join(ansibleDir, GLOBAL_GROUP_VARS_DIR)) {
		fmt.Println("GetAnsibleFileConfig ansibleDir=", ansibleDir)
	}
	ansibleGlobalDir := path.Join(ansibleDir, GLOBAL_GROUP_VARS_DIR)

	srvsCreds, err := getGetAnsibleSshCreds(ansibleGlobalDir)
	if err != nil || len(srvsCreds) < 1 {
		t.Fatalf(`TestGetGetAnsibleSshCreds Failed, error=%v`, err)
	}

}

func TestGetAnsibleFileConfig(t *testing.T) {
	ansibleDir := ""
	GLOBAL_GROUP_VARS_DIR := "00.global/group_vars/all"

	if ansibleDir = "../../test/ansible-inventory/CLASSIC"; exists(path.Join(ansibleDir, GLOBAL_GROUP_VARS_DIR)) {
		fmt.Println("GetAnsibleFileConfig ansibleDir=", ansibleDir)
	}
	config, err := GetAnsibleFileConfig(ansibleDir, "30.")
	if err != nil {
		t.Fatalf(`TestGetAnsibleFileConfig Failed, error=%v`, err)
	}
	if len(config.Environments) < 2 {
		t.Fatalf(`TestGetAnsibleFileConfig Failed, config.Environments<2`)
	}
}
