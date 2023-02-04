package tasks

import "serversBulk/modules/configProvider"

type TaskType string

const (
	TypeUploadFile      TaskType = "UploadFile"
	TypeArchiveLogs     TaskType = "ArchiveLogs"
	TypeArchiveGzip     TaskType = "TypeArchiveGzip"
	TypeDownloadArchive TaskType = "DownloadLogs"
	TypeGrepInLogs      TaskType = "GrepInLogs"
	TypeExecuteCommand  TaskType = "ExecuteCommand"
)

type TaskStatus string

const (
	Planned    TaskStatus = "Planned"
	InProgress TaskStatus = "InProgress"
	Finished   TaskStatus = "Finished"
	Failed     TaskStatus = "Failed"
)

type ServerTask struct {
	Type           TaskType
	Status         TaskStatus
	RemoteFileName string
	Log            string
	ConfigFileName string
	ServersName    string
	ModifTime      string
	GrepFor        string
	ExecuteCmd     string
	//LocalDir - for dwnloading files
	LocalDir       string
	LocalFile      string
	LogFilePattern string
	ConfigServer   configProvider.ConfigServerType
	Server         string
	Error          error
}
