package main

import (
	"fmt"

	"sebulk/modules/sshHelper"
	"sebulk/modules/tasks"
)

// NOTE: output result is in the channel
func grepInLogs(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	// TODO: improve serach by showing only data for specified period:
	statusHandler(task.Server, "CONNECTING...")
	logHandler(task.Server, fmt.Sprintf("connecting to server: [%s] to grep", task.Server))
	strOutput := ""
	useMtime := false
	// var mHours int
	// var curSrvTime, timeLogs time.Time
	// t := time.Now()
	sshAdv, err := sshHelper.OpenSshAdvanced(&task.ConfigServer, task.Server)
	if err == nil {
		defer sshAdv.Close()
		// if task.ModifTime != "" {
		// 	var mDays float64
		// 	mDays, err = strconv.ParseFloat(task.ModifTime, 64)
		// 	if err == nil {
		// 		// mHours = int(mDays * 24)
		// 		useMtime = true
		// 	}
		// }
		// get from Linux: date "+%Y-%m-%d %H:%M:%S" !!!
		// if err == nil && useMtime {
		// 	strOutput, err = executeWithConnection(sshAdv, task.Server, `date -u +"%Y-%m-%dT%H:%M:%S"`)
		// 	if err == nil {
		// 		curSrvTime, err = time.Parse(time.RFC3339, strOutput)
		// 		if err == nil {
		// 			timeLogs = curSrvTime.Add(time.Duration(mHours) * time.Hour)
		// 		}
		// 	}
		// }
		if err == nil {
			// 2022.13.08 20\:46\:36.824
			// 2022.14.08 20:21:58.270
			// 14.08.2022 20\:22\:06.816
			// 15.08.2022 01:30:35.202
			// 12.08.2022 02:14:35
			// 2022-08-12	15:42:09
			// Aug 4, 2022 6:11:05 - [A-Za-z]{3}\s\d{1,2},\s2\d{3}

			//USE AWK instead of Grep:
			// awk  '/Aug 4, 2022 6\:11\:05.*(Error|ERROR)/ {print $0; f = 1;
			// if (match($0,/Aug 4, 2022/)) printf "MATCHED:[%s]", substr($0, RSTART, RLENGTH); next} f;/Aug 4, 2022/ {if (f == 1){f = 0; print "#######################"};}' ./admin.log

			// AWK SIMPLE: awk  '/Aug 4, 2022 6\:11\:05.*(Error|ERROR)/{ print $0; f = 1; next } f; /Aug 4, 2022/ { if (f == 1){ f = 0; print "#######################"}}' ./admin.log
			// Parse in GO ?

			// AWK with seraching date
			// # DATE:
			// [A-Za-z]{3} [0-9]{1,2}, 2[0-9]{3} [0-9]{1,2}\:[0-9]{2}\:[0-9]{2}
			// awk  '/[A-Za-z]{3} [0-9]{1,2}, 2[0-9]{3} [0-9]{1,2}\:[0-9]{2}\:[0-9]{2}[^\n]+(Error|ERROR)/{ print $0; f = 1; if (match($0,/[A-Za-z]{3} [0-9]{1,2}, 2[0-9]{3} [0-9]{1,2}\:[0-9]{2}\:[0-9]{2}/)) printf "MATCHED:[%s]", substr($0, RSTART, RLENGTH);next } f; /[A-Za-z]{3} [0-9]{1,2}, 2[0-9]{3} [0-9]{1,2}\:[0-9]{2}\:[0-9]{2}/ { if (f == 1){ f = 0; print "+++++++++++"}}' ./admin.log > ~/Downloads/zout3.txt

			strGrep := fmt.Sprintf("grep --color=auto -H -A25 -B3 -i \"%s\" {}  \\;", task.CommandCargo)
			strMTime := ""
			if useMtime {
				strMTime = fmt.Sprintf("-mtime %s", task.ModifTime)
			}
			task.ExecuteCmd = "cd ~"
			for _, folder := range task.ConfigServer.LogFolders {
				task.ExecuteCmd = fmt.Sprintf("%s; find %s ! -readable -prune -o -type f -iname \"%s\" %s -exec %s", task.ExecuteCmd, folder, task.ConfigServer.LogFilePattern, strMTime, strGrep)
			}
			strOutput, err = executeWithConnection(sshAdv, task.Server, task.ExecuteCmd)
		}
	}
	output <- *taskForChannel(&task, strOutput, err, tasks.Finished, nil)
}
