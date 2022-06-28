package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"serversBulk/modules/configProvider"
	"serversBulk/modules/logHelper"
	"serversBulk/modules/sshHelper"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
)

type CurrentFileForDownloading struct {
	Server       string
	FileName     string
	ConfigServer configProvider.ConfigServerType
	SshConfig    *ssh.ClientConfig
	Error        error
}

func main() {
	logHelper.LogPrintln("******* START serversBulk *********")
	// fmt.Println("The tool for quick serach in the logs of different nodes\nand servers when Grafana is not available")
	// fmt.Println("EXAMPLES:\n\tto search in logs:\n\t\t./serversBulk -s \"search\" > ./output.txt")
	// fmt.Println("\t\t./serversBulk -c ./serversBulk_config_SVT.json -s \"18\\:.*\\[ERROR\"")
	// fmt.Println("\n\tto download logs:\n\t\t./serversBulk -d ./logs/")

	jsonFileName := flag.String("c", "./config/serversBulk_config.json", "path to environment configuration file")
	serversName := flag.String("servers", "", "to search/download only from the servers with NAME='servers', \n\tfor example if you need to download from TBAPI\n\tserevers you can use parameter: `--servers TBAPI` ")
	modifTime := flag.String("mtime", "-0.2", "same as mtime for 'find'")
	grepFor := flag.String("s", "", "search string like in:\ngrep --color=auto -H -A2 -B4  \"search string\"")
	executeCmd := flag.String("e", "", "execute given command:\nserversBulk --servers TBAPI -e \"curl -v -g http://localhost:28080/api/v1/monitoring/health\"\n\tto get TBAPI health from all TBAPI nodes")
	logDir := flag.String("d", "", "folder where log files should be downloaded")
	uploadLocalFile := flag.String("u", "", "File which will be uploaded to /var/tmp to the target servers")
	logFilePattern := flag.String("f", "", "log File pattern: i.e. *.log the value will overwrite value in config")
	flag.Parse()
	if *grepFor == "" && *logDir == "" && *executeCmd == "" && *uploadLocalFile == "" {
		logHelper.ErrFatal(errors.New("-s or -d or -e or -u should be specified"))
	} else if *grepFor != "" && *logDir != "" {
		logHelper.ErrFatal(errors.New("only one of these: -s or -d should be specified"))
	}

	config := configProvider.GetConfig(jsonFileName)
	// Calculate number of channels
	numberOfChannels := 0
	for _, serverConf := range config.Servers {
		if serverConf.IpAddresses != nil && (*serversName == "" || serverConf.Name == *serversName) {
			numberOfChannels += len(serverConf.IpAddresses)
		}
	}
	logHelper.LogPrintf("calculated number of channels/servers to connect=[%v]", numberOfChannels)
	if numberOfChannels < 1 {
		logHelper.ErrFatal(errors.New("servers to connect is less than 1 - nothing to do"))
	}

	resultChannels := make(chan string)                     //, numberOfChannels)
	zipFilesChannel := make(chan CurrentFileForDownloading) //, numberOfChannels)
	for _, serverConf_p := range config.Servers {
		serverConf := serverConf_p
		if serverConf.IpAddresses != nil && (*serversName == "" || serverConf.Name == *serversName) {
			if *logFilePattern != "" {
				serverConf.LogFilePattern = *logFilePattern
			}
			for _, serverIp := range serverConf.IpAddresses {
				if *executeCmd != "" {
					go executeOnServerWithOutput(&serverConf, serverIp, *executeCmd, resultChannels)
				} else if *grepFor != "" {
					go grepOnServer(&serverConf, serverIp, *grepFor, *modifTime, resultChannels)
				} else if *uploadLocalFile != "" {
					go uploadFile(&serverConf, serverIp, *uploadLocalFile, "/var/tmp", resultChannels)
				} else {
					go prepareZipLog(&serverConf, serverIp, *modifTime, zipFilesChannel)
					go downloadZipLog(&serverConf, *logDir, zipFilesChannel, resultChannels)
				}
			}
		}
	}

	// logPrintln("Waiting for complition")
	// wg.Wait()

	// fmt.Printf("Result: %s \n", results)
	count := 0
	for i := range resultChannels {
		count++
		fmt.Println("")
		fmt.Println("")
		logHelper.LogPrintf("count: %v of %v", count, numberOfChannels)
		fmt.Printf("Result: %v of %v\n", count, numberOfChannels)
		fmt.Println("↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓")
		fmt.Print(i)
		fmt.Println("")
		fmt.Println("↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑")
		if count >= numberOfChannels {
			break
		}
	}

	//Species: pigeon, Description: likes to perch on rocks
	logHelper.LogPrintln("FINISHED SUCCESFULLY")
}

// NOTE: output result is in the channel
func grepOnServer(serverConf *configProvider.ConfigServerType, server, grepFor, modifTime string, output chan<- string) {
	cmd := fmt.Sprintf("find %s -type f -iname \"%s\" -mtime %s -exec grep --color=auto -H -A15 -B15 \"%s\" {}  \\;", serverConf.LogFolder, serverConf.LogFilePattern, modifTime, grepFor)
	executeOnServerWithOutput(serverConf, server, cmd, output)
}
func executeOnServer(serverConf *configProvider.ConfigServerType, server, cmd string) (string, error) {
	logHelper.LogPrintf("executeOnServer server: %s", server)
	// strOutut = make(chan string)
	sshAdv := sshHelper.OpenSshAdvanced(serverConf, server)
	defer sshAdv.Close()
	logHelper.LogPrintf("execute command:[%s] on ssh server:[%s]", cmd, server)
	buff, e := sshAdv.NewSession().Output(cmd)
	if e != nil {
		logHelper.ErrLogWinMessage(fmt.Sprintf("error while executing cmd:[%s] os server [%s], cmd_output[%s]", cmd, server, buff), e)
	}
	str := fmt.Sprintf("\nSERVER:%s\n_____________________________\n%s", color.HiYellowString(server), string(buff))
	return str, e
}
func executeOnServerWithOutput(serverConf *configProvider.ConfigServerType, server, cmd string, output chan<- string) {
	str, _ := executeOnServer(serverConf, server, cmd)
	output <- str
}

func prepareZipLog(serverConf *configProvider.ConfigServerType, server, modifTime string, outZipFilename chan<- CurrentFileForDownloading) {

	tarNamefile := fmt.Sprintf("%s/%s.%s", serverConf.LogFolder, strings.ReplaceAll(server, ".", "_"), "tar")
	cmd := fmt.Sprintf("cd %s; find ./ -type f -iname \"%s\" -mtime %s -exec tar rvf %s {} \\;", serverConf.LogFolder, serverConf.LogFilePattern, modifTime, tarNamefile)
	_, e := executeOnServer(serverConf, server, cmd)
	if e == nil {
		tarGzNamefile := fmt.Sprintf("%s.gz", tarNamefile)

		cmdGz := fmt.Sprintf("cd %s; tar cvzf %s %s ; rm %s", path.Dir(tarGzNamefile), path.Base(tarGzNamefile), path.Base(tarNamefile), tarNamefile)
		if _, e = executeOnServer(serverConf, server, cmdGz); e == nil {
			outZipFilename <- CurrentFileForDownloading{FileName: tarGzNamefile, Server: server, ConfigServer: *serverConf}
		} else {
			outZipFilename <- CurrentFileForDownloading{Server: server, Error: e}
		}
	} else {
		outZipFilename <- CurrentFileForDownloading{Server: server, Error: e}
	}
}

func downloadZipLog(serverConf *configProvider.ConfigServerType, localDir string, inZipFilename chan CurrentFileForDownloading, output chan<- string) {
	zipFileSrv := <-inZipFilename
	if zipFileSrv.Error == nil {
		zipFilePath := zipFileSrv.FileName
		server := zipFileSrv.Server
		zipFileName := path.Base(zipFilePath)
		localZipFileName := path.Join(localDir, zipFileName)

		logHelper.LogPrintf("Downloading File=[%s] from server=[%s] to local path=[%s] ...", zipFilePath, server, localZipFileName)

		sshAdv := sshHelper.OpenSshAdvanced(serverConf, server)
		defer sshAdv.Close()
		sftpClient := sshAdv.NewSftpClient()
		srcFile, err := sftpClient.OpenFile(zipFilePath, (os.O_RDONLY))
		if err != nil {
			logHelper.ErrFatalWithMessage(
				fmt.Sprintf("Unable to open file=[%s] on server=[%s]", zipFilePath, server),
				err)
		}

		dstFile, err := os.Create(localZipFileName)
		if err != nil {
			logHelper.ErrFatal(err)
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			logHelper.ErrFatal(err)
		}

		logHelper.LogPrintf("removing file:%s\n", zipFilePath)
		err = sftpClient.Remove(zipFilePath)
		if err != nil {
			logHelper.ErrFatal(err)
		}

		output <- fmt.Sprintf("Created File: %s source server: %s", localZipFileName, server)
	} else {
		output <- fmt.Sprintf("Server: %s finished with Error: %s", zipFileSrv.Server, zipFileSrv.Error)
	}
}

func uploadFile(serverConf *configProvider.ConfigServerType, server string, localFile string, destFolder string, output chan<- string) {
	destFilePath := path.Join(destFolder, path.Base(localFile))

	logHelper.LogPrintf("Uploading File=[%s] to server=[%s] to dest path=[%s] ...", localFile, server, destFilePath)

	sshAdv := sshHelper.OpenSshAdvanced(serverConf, server)
	defer sshAdv.Close()
	sftpClient := sshAdv.NewSftpClient()
	dstFile, err := sftpClient.Create(destFilePath)
	if err != nil {
		logHelper.ErrFatalWithMessage(
			fmt.Sprintf("Unable to create file=[%s] on server=[%s]", destFilePath, server),
			err)
	}

	srcFile, err := os.Open(localFile)
	if err != nil {
		logHelper.ErrFatal(err)
	}
	defer srcFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		logHelper.ErrFatal(err)
	}
	output <- fmt.Sprintf("Created File: %s target server: %s", destFilePath, server)
}
