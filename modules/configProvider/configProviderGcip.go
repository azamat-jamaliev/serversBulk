package configProvider

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
)

// exists returns whether the given file or directory exists
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

type SshCreds struct {
	Login, Password string
}

func getGetAnsibleSshCreds() (map[string]SshCreds, error) {
	ansibleDir := ""
	GLOBAL_GROUP_VARS_DIR := "00.global/group_vars/all"
	result := map[string]SshCreds{}

	if ansibleDir = "./ansible-inventory/CLASSIC"; exists(path.Join(ansibleDir, GLOBAL_GROUP_VARS_DIR)) {
		fmt.Println("GetAnsibleFileConfig ansibleDir=", ansibleDir)
	} else if ansibleDir = "./CLASSIC"; exists(path.Join(ansibleDir, GLOBAL_GROUP_VARS_DIR)) {
		fmt.Println("GetAnsibleFileConfig ansibleDir=", ansibleDir)
	}
	ansibleGlobalDir := path.Join(ansibleDir, GLOBAL_GROUP_VARS_DIR)
	globalDirFiles, err := os.ReadDir(ansibleGlobalDir)
	if err != nil {
		return nil, err
	}
	for _, f := range globalDirFiles {
		if !f.IsDir() {
			fi, err := f.Info()
			if err != nil {
				return nil, err
			}
			filename := fi.Name()
			ext := path.Ext(filename)
			serverName := filename[0 : len(filename)-len(ext)]

			yamlFile, err := os.Open(path.Join(ansibleGlobalDir, f.Name()))
			if err != nil {
				return nil, err
			}
			bytes, _ := ioutil.ReadAll(yamlFile)
			yamlFileContent := string(bytes)

			searchSshLoginString := fmt.Sprintf(`(?i)%s_ansible_ssh_user:\s*['"]([^"']+)`, serverName)
			fmt.Println("searchSshLoginString: ", searchSshLoginString)
			reLogin := regexp.MustCompile(searchSshLoginString)
			matchesLogin := reLogin.FindStringSubmatch(yamlFileContent)

			searchSshPasswordString := fmt.Sprintf(`(?i)%s_ansible_ssh_pass:\s*['"]([^"']+)`, serverName)
			fmt.Println("searchSshPasswordString: ", searchSshPasswordString)
			rePassword := regexp.MustCompile(searchSshPasswordString)
			matchesPassword := rePassword.FindStringSubmatch(yamlFileContent)

			if len(matchesPassword) > 1 && len(matchesLogin) > 1 {
				fmt.Println("Found SSH Login: ", matchesLogin[1])
				fmt.Println("Found SSH password:", matchesPassword[1])
				result[serverName] = SshCreds{Login: matchesLogin[1], Password: matchesPassword[1]}
				// return matchesLogin[1], matchesPassword[1], nil
			}
		}
	}
	return result, nil
}

// func GetAnsibleFileConfig() (ConfigFileType, error) {
// 	var config ConfigFileType
// 	ansibleDir := ""
// 	GLOBAL_GROUP_VARS_DIR := "00.global/group_vars/all"
// 	ENV_GROUP_DIR_PREFIX := "30."
// 	if ansibleDir = "./ansible-inventory/CLASSIC"; exists(path.Join(ansibleDir, GLOBAL_GROUP_VARS_DIR)) {
// 		fmt.Println("GetAnsibleFileConfig ansibleDir=", ansibleDir)
// 	} else if ansibleDir = "./CLASSIC"; exists(path.Join(ansibleDir, GLOBAL_GROUP_VARS_DIR)) {
// 		fmt.Println("GetAnsibleFileConfig ansibleDir=", ansibleDir)
// 	}
// 	ansibleGlobalDir := path.Join(ansibleDir, GLOBAL_GROUP_VARS_DIR)
// 	globalDirFiles, err := os.ReadDir(ansibleGlobalDir)
// 	if err != nil {
// 		return config, err
// 	}
// 	for _, f := range globalDirFiles {
// 		if !f.IsDir() {
// 			fi, err := f.Info()
// 			if err != nil {
// 				return config, err
// 			}
// 			searchSshLoginString := fmt.Sprintf(`%s_ansible_ssh_user:\s*['"]([^"']+)`, fi.Name())
// 			var reLogin = regexp.MustCompile(searchSshLoginString)
// 			jsonFile, err := os.Open(path.Join(ansibleGlobalDir, f.Name()))
// 			if err != nil {
// 				return config, err
// 			}
// 			bytes, _ := ioutil.ReadAll(jsonFile)
// 			// err = json.Unmarshal(jsonFileBytes, &config)
// 			matchesLogin := reLogin.FindStringSubmatch(string(bytes))
// 			fmt.Printf("%s", matchesLogin[1])

// 		}
// 	}

// 	dirs, err := os.ReadDir(ansibleDir)
// 	if err != nil {
// 		return config, err
// 	}
// 	for _, dir := range dirs {
// 		if dir.IsDir() {
// 			fi, err := dir.Info()
// 			if err != nil {
// 				return config, err
// 			}
// 			prefix := fi.Name()[0:len(ENV_GROUP_DIR_PREFIX)]
// 			if prefix == ENV_GROUP_DIR_PREFIX {
// 				fmt.Println("Reading folder:", fi.Name())
// 			}
// 		}
// 	}

// 	return config, err
// }
