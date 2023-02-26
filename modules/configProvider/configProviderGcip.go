package configProvider

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
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

func getValueFromYamlFile(attrName, yamlFileContent string) string {
	val := getMatchingValueFromYamlFile(fmt.Sprintf(`(?i)%s:\s*['"]([^"']+)`, attrName), yamlFileContent)
	if len(val) < 1 {
		val = getMatchingValueFromYamlFile(fmt.Sprintf(`(?i)%s:\s*([^\s\n]+)`, attrName), yamlFileContent)
	}
	return val
}
func getMatchingValueFromYamlFile(searchMatching, yamlFileContent string) string {
	fmt.Println("getMatchingValueFromYamlFile yaml serach matching: ", searchMatching)
	reLogin := regexp.MustCompile(searchMatching)
	matchesLogin := reLogin.FindStringSubmatch(yamlFileContent)
	if len(matchesLogin) > 1 {
		return matchesLogin[1]
	}
	return ""
}
func getServerTypeNameFromFile(filename string) string {
	ext := path.Ext(filename)
	return strings.ToLower(filename[0 : len(filename)-len(ext)])
}
func getGetAnsibleSshCreds(ansibleGlobalDir string) (map[string]SshCreds, error) {
	result := map[string]SshCreds{}

	globalDirFiles, err := os.ReadDir(ansibleGlobalDir)
	if err != nil {
		return nil, err
	}
	for _, f := range globalDirFiles {
		if !f.IsDir() {
			serverName := getServerTypeNameFromFile(f.Name())

			yamlFile, err := os.Open(path.Join(ansibleGlobalDir, f.Name()))
			if err != nil {
				return nil, err
			}
			bytes, _ := ioutil.ReadAll(yamlFile)
			yamlFileContent := string(bytes)
			login := getValueFromYamlFile(fmt.Sprintf("%s_ansible_ssh_user", serverName), yamlFileContent)
			pass := getValueFromYamlFile(fmt.Sprintf("%s_ansible_ssh_user", serverName), yamlFileContent)

			if len(login) > 1 && len(pass) > 1 {
				result[serverName] = SshCreds{Login: login, Password: pass}
			}
		}
	}
	return result, nil
}

func GetAnsibleFileConfig(ansibleDir, envPrefix string) (ConfigFileType, error) {
	config := GetDefaultConfig()
	creds, err := getGetAnsibleSshCreds(path.Join(ansibleDir, "00.global/group_vars/all"))
	if err != nil {
		return config, err
	}

	dirs, err := os.ReadDir(ansibleDir)
	if err != nil {
		return config, err
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			prefix := dir.Name()[0:len(envPrefix)]
			if prefix == envPrefix && len(dir.Name()) > len(prefix) {
				fmt.Println("Reading folder:", dir.Name())
				env := ConfigEnvironmentType{
					Name:    dir.Name()[len(envPrefix) : len(dir.Name())-len(envPrefix)],
					Servers: []ConfigServerType{},
				}

				envHostDir := path.Join(ansibleDir, dir.Name(), "host_vars")
				files, err := os.ReadDir(envHostDir)
				if err != nil {
					return config, err
				}
				for _, f := range files {
					if !f.IsDir() {
						serverName := getServerTypeNameFromFile(f.Name())
						if cred, ok := creds[serverName]; ok {
							yamlFile, err := os.Open(path.Join(envHostDir, f.Name()))
							if err != nil {
								return config, err
							}
							bytes, _ := ioutil.ReadAll(yamlFile)
							srvIp := getValueFromYamlFile("ansible_host", string(bytes))
							if len(srvIp) > 0 {
								serv := ConfigServerType{
									Name:        serverName,
									IpAddresses: []string{srvIp},
									Login:       cred.Login,
									Passowrd:    cred.Password,
								}
								env.Servers = append(env.Servers, serv)
							}
						}
					}
				}
				if len(env.Servers) > 0 {
					config.Environments = append(config.Environments, env)
				}
			}
		}
	}

	return config, err
}
