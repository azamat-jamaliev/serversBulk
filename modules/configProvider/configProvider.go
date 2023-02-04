package configProvider

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type ConfigServerType struct {
	Name                string
	LogFolder           string
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
	DownloadFolder      string
	LogsMtime           *float32
	LogFolder           string
	LogFilePattern      string
	Login               string
	Passowrd            string
	IdentityFile        string
	BastionServer       string
	BastionLogin        string
	BastionIdentityFile string
	BastionPassword     string
	Environments        []ConfigEnvironmentType
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
func GetFileConfig(jsonFileName *string) ConfigFileType {
	jsonFile, err := os.Open(*jsonFileName)
	if err != nil {
		panic(err)
	}

	var config ConfigFileType
	jsonFileBytes, _ := ioutil.ReadAll(jsonFile)
	if err := json.Unmarshal(jsonFileBytes, &config); err != nil {
		panic(err)
	}
	return config
}
