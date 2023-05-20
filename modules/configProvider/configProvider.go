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
	Name      string //environment name
	DoNotSave bool
	Servers   []ConfigServerType
}
type ConfigFileType struct {
	DownloadFolder string
	UploadFolder   string
	LogsMtime      *float64
	Environments   []ConfigEnvironmentType
}

func (t *ConfigFileType) GetEnvironmentByName(envName string) *ConfigEnvironmentType {
	for _, env := range t.Environments {
		if env.Name == envName {
			return &env
		}
	}
	return nil
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
	modifiedConf := conf
	modifiedConf.Environments = []ConfigEnvironmentType{}
	for _, env := range conf.Environments {
		if !env.DoNotSave {
			modifiedConf.Environments = append(modifiedConf.Environments, env)
		}
	}
	bytes, err := json.Marshal(modifiedConf)
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
				Name: "EXAMPLE_environemnt",
				Servers: []ConfigServerType{{
					Name:           "example_servers_group1",
					IpAddresses:    []string{"127.0.0.1"},
					LogFolders:     []string{"~/"},
					LogFilePattern: "*.log*",
					Login:          "test",
					Passowrd:       "test",
				}},
			}},
	}
}
