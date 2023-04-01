package tasks

import "sebulk/modules/configProvider"

type TaskType string

const (
	TypeUploadFile      TaskType = "UploadFile"
	TypeArchiveLogs     TaskType = "ArchiveLogs" //this is first task in Download Logs sequence
	TypeArchiveGzip     TaskType = "TypeArchiveGzip"
	TypeDownloadArchive TaskType = "DownloadLogs"
	TypeGrepInLogs      TaskType = "GrepInLogs"
	TypeAwkInLogs       TaskType = "AwkInLogs"
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
	ModifTime      string
	CommandCargo   string
	CommandCargo2  string
	ExecuteCmd     string
	LocalFile      string
	// GrepFor        string
	// LocalDir - for dwnloading files
	// LocalDir       string
	// LogFilePattern string
	ConfigServer configProvider.ConfigServerType
	Server       string
	Error        error
}
