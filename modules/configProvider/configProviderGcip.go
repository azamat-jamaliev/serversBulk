package configProvider

import (
	"fmt"
	"io"
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

type SshServerYamlDetails struct {
	Login      string
	Password   string
	LogFolders []string
}

func getValueFromYamlFile(attrName, yamlFileContent string) string {
	val := getMatchingValueFromYamlFile(fmt.Sprintf(`(?i)%s:\s*['"]([^"']+)`, attrName), yamlFileContent)
	if len(val) < 1 {
		val = getMatchingValueFromYamlFile(fmt.Sprintf(`(?i)%s:\s*([^\s\n'"]+)`, attrName), yamlFileContent)
	}
	return val
}
func getMatchingValueFromYamlFile(searchMatching, yamlFileContent string) string {
	reLogin := regexp.MustCompile(searchMatching)
	matchesLogin := reLogin.FindStringSubmatch(yamlFileContent)
	if len(matchesLogin) > 1 {
		return matchesLogin[1]
	}

	return ""
}
func getServerTypeNameFromFile(filename string) string {
	ext := path.Ext(filename)
	return strings.ToLower(filename[:len(filename)-len(ext)])
}
func appendIfNotEmpty(slice []string, value string) []string {
	if value != "" {
		return append(slice, value)
	}
	return slice
}
func getGetAnsibleSshServerYamlDetails(ansibleGlobalDir string) (map[string]SshServerYamlDetails, error) {
	result := map[string]SshServerYamlDetails{}

	globalDirFiles, err := os.ReadDir(ansibleGlobalDir)
	if err == nil {
		countLogs := 0
		for _, f := range globalDirFiles {
			if !f.IsDir() {
				serverName := getServerTypeNameFromFile(f.Name())

				yamlFile, err := os.Open(path.Join(ansibleGlobalDir, f.Name()))
				if err != nil {
					return nil, err
				}
				bytes, _ := io.ReadAll(yamlFile)
				yamlFileContent := string(bytes)
				login := getValueFromYamlFile(fmt.Sprintf("%s_ansible_ssh_user", serverName), yamlFileContent)
				pass := getValueFromYamlFile(fmt.Sprintf("%s_ansible_ssh_pass", serverName), yamlFileContent)

				logs := []string{}
				logs = appendIfNotEmpty(logs, getValueFromYamlFile(fmt.Sprintf("%s_logs_folder", serverName), yamlFileContent))
				logs = appendIfNotEmpty(logs, getValueFromYamlFile(fmt.Sprintf("%s_installer_logs_folder", serverName), yamlFileContent))

				if login != "" && pass != "" {
					result[serverName] = SshServerYamlDetails{Login: login, Password: pass, LogFolders: logs}
				}
				if len(logs) > 0 {
					countLogs++
				}
			}
		}
		if countLogs == 0 {
			err = fmt.Errorf("ERROR: no log folders like: [server_logs_folder] or [server_installer_logs_folder] were found in yaml files in: [%s]", globalDirFiles)
		}
	}
	return result, err
}

func GetAnsibleEnvironmentsConfig(ansibleDir, envPrefix string) ([]ConfigEnvironmentType, error) {
	envs := []ConfigEnvironmentType{}
	creds, err := getGetAnsibleSshServerYamlDetails(path.Join(ansibleDir, "00.global/group_vars/all"))
	if err != nil {
		return nil, err
	}

	dirs, err := os.ReadDir(ansibleDir)
	if err != nil {
		return nil, err
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			prefix := dir.Name()[:len(envPrefix)]
			if prefix == envPrefix && len(dir.Name()) > len(prefix) {
				env := ConfigEnvironmentType{
					Name:      dir.Name()[len(envPrefix):],
					Servers:   []ConfigServerType{},
					DoNotSave: true,
				}

				envHostDir := path.Join(ansibleDir, dir.Name(), "host_vars")
				files, err := os.ReadDir(envHostDir)
				if err != nil {
					return envs, err
				}
				for _, f := range files {
					if !f.IsDir() {
						serverName := getServerTypeNameFromFile(f.Name())
						if cred, ok := creds[serverName]; ok {
							yamlFile, err := os.Open(path.Join(envHostDir, f.Name()))

							if err != nil {
								return envs, err
							}
							bytes, _ := io.ReadAll(yamlFile)
							srvIp := getValueFromYamlFile("ansible_host", string(bytes))
							if len(srvIp) > 0 {
								serv := ConfigServerType{
									Name:           serverName,
									IpAddresses:    []string{srvIp},
									Login:          cred.Login,
									Passowrd:       cred.Password,
									LogFolders:     cred.LogFolders,
									LogFilePattern: "*.log*",
								}
								env.Servers = append(env.Servers, serv)
							}
						}
					}
				}
				if len(env.Servers) > 0 {
					envs = append(envs, env)
				}
			}
		}
	}

	return envs, err
}
