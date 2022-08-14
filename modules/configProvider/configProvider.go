package configProvider

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"serversBulk/modules/logHelper"
)

type ConfigServerType struct {
	Name                string
	Description         string
	LogFolder           string
	LogFilePattern      string
	SearchInSubfolders  bool
	Login               string
	Passowrd            string
	BastionServer       string
	BastionLogin        string
	BastionIdentityFile string
	BastionPassword     string
	IpAddresses         []string
}
type ConfigFileType struct {
	Servers []ConfigServerType
}

func GetConfig(jsonFileName *string) ConfigFileType {
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
