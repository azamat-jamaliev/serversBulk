package configProvider

import (
	"encoding/json"
	"io"
	"os"
)

type ConfigServerType struct {
	Name                string
	LogFolders          []string
	LogFilePattern      string
	Login               string
	Passowrd            string
	IdentityFile        string
	BastionServer       string
	BastionLogin        string
	BastionIdentityFile string
	BastionPassword     string
	IpAddresses         []string
}
type ConfigEnvironmentType struct {
	Name    string //environment name
	Servers []ConfigServerType
}
type ConfigFileType struct {
	DownloadFolder string
	UploadFolder   string
	LogsMtime      *float64
	Environments   []ConfigEnvironmentType
}

func GetEnvironemntConfig(jsonFileName *string) ConfigEnvironmentType {
	jsonFile, err := os.Open(*jsonFileName)
	if err != nil {
		panic(err)
	}

	var config ConfigEnvironmentType
	jsonFileBytes, _ := io.ReadAll(jsonFile)
	if err := json.Unmarshal(jsonFileBytes, &config); err != nil {
		panic(err)
	}
	return config
}
func GetFileConfig(jsonFileName string) (ConfigFileType, error) {
	var config ConfigFileType
	jsonFile, err := os.Open(jsonFileName)
	if err == nil {
		jsonFileBytes, _ := io.ReadAll(jsonFile)
		err = json.Unmarshal(jsonFileBytes, &config)
	}
	return config, err
}
func SaveFileConfig(jsonFileName *string, conf ConfigFileType) {
	bytes, err := json.Marshal(conf)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(*jsonFileName, bytes, 0644); err != nil {
		panic(err)
	}
}
func GetDefaultConfig() ConfigFileType {
	f := -0.2
	return ConfigFileType{
		DownloadFolder: ".",
		UploadFolder:   "/var/tmp",
		LogsMtime:      &f,
		Environments: []ConfigEnvironmentType{
			{
				Name: "EXAMPLE_local_test",
				Servers: []ConfigServerType{{
					Name:           "sebulk_test_ubuntu",
					IpAddresses:    []string{"127.0.0.1"},
					LogFolders:     []string{"/var/log"},
					LogFilePattern: "*.log*",
					Login:          "test",
					Passowrd:       "test",
				}},
			}},
	}
}
