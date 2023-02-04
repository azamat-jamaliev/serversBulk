package configProvider

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"serversBulk/modules/logHelper"
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
	LogsMtime           float32
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
	logHelper.LogPrintf("loading environment configuration file [%s]", *jsonFileName)
	jsonFile, err := os.Open(*jsonFileName)
	if err != nil {
		logHelper.ErrFatalln(err, "Cannot Open Environment file")
	}

	var config ConfigEnvironmentType
	jsonFileBytes, _ := ioutil.ReadAll(jsonFile)
	if err := json.Unmarshal(jsonFileBytes, &config); err != nil {
		logHelper.ErrFatalln(err, "Cannot Parse Environment file")
	}
	logHelper.LogPrintf("Environment config loaded succesfully")
	return config
}
func GetFileConfig(jsonFileName *string) ConfigFileType {
	logHelper.LogPrintf("loading configuration file [%s]", *jsonFileName)
	jsonFile, err := os.Open(*jsonFileName)
	if err != nil {
		logHelper.ErrFatalln(err, "Cannot Open Config file")
	}

	var config ConfigFileType
	jsonFileBytes, _ := ioutil.ReadAll(jsonFile)
	if err := json.Unmarshal(jsonFileBytes, &config); err != nil {
		logHelper.ErrFatalln(err, "Cannot Parse Config file")
	}
	logHelper.LogPrintf("config loaded succesfully")
	return config
}
