package configProvider

import (
	"encoding/json"
	"io/ioutil"
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
	LogsMtime      *float64
	Environments   []ConfigEnvironmentType
}

func GetEnvironemntConfig(jsonFileName *string) ConfigEnvironmentType {
	jsonFile, err := os.Open(*jsonFileName)
	if err != nil {
		panic(err)
	}

	var config ConfigEnvironmentType
	jsonFileBytes, _ := ioutil.ReadAll(jsonFile)
	if err := json.Unmarshal(jsonFileBytes, &config); err != nil {
		panic(err)
	}
	return config
}
func GetFileConfig(jsonFileName string) (ConfigFileType, error) {
	var config ConfigFileType
	jsonFile, err := os.Open(jsonFileName)
	if err == nil {
		jsonFileBytes, _ := ioutil.ReadAll(jsonFile)
		err = json.Unmarshal(jsonFileBytes, &config)
	}
	return config, err
}
func SaveFileConfig(jsonFileName *string, conf ConfigFileType) {
	bytes, err := json.Marshal(conf)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(*jsonFileName, bytes, 0644); err != nil {
		panic(err)
	}
}
