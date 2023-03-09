package main

import (
	"fmt"
	"strings"

	"sebulk/modules/sshHelper"
	"sebulk/modules/tasks"
)

// NOTE: output result is in the channel
func awkInLogs(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	// TODO: improve serach by showing only data for specified period:
	statusHandler(task.Server, "CONNECTING...")
	logHandler(task.Server, fmt.Sprintf("connecting to server: [%s] to awk", task.Server))
	strOutput := ""
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
			// ^.{0,10}[0-9]{1,2}[-\.][0-9]{1,2}[-\.]2[0-9]{3} [0-9]{1,2}[\\:]{1,2}[0-9]{2}
			dateRegExp := []string{
				`^.{0,10}[0-9]{1,2}[-\.][0-9]{1,2}[-\.]2[0-9]{3} [0-9]{1,2}[\\:]{1,2}[0-9]{2}`,
				`^.{0,10}2[0-9]{3}[-\.][0-9]{1,2}[-\.][0-9]{1,2} [0-9]{1,2}[\\:]{1,2}[0-9]{2}`,
				`^.{0,10}[A-Za-z]{3} [0-9]{1,2}, 2[0-9]{3} [0-9]{1,2}:[0-9]{2}`,
			}

			strRegExDates := strings.Join(dateRegExp, "|")
			strAwk := fmt.Sprintf("awk  '/(%s)[^\\n]{0,60}(Error|ERROR)/{ print $0; f = 1 ;next } f; /(%s)/ { if (f == 1){ f = 0; print \"+++++++++++\"}}' {} ", strRegExDates, strRegExDates)
			task.ExecuteCmd = getFindExecForTask(task, strAwk)
			strOutput, err = executeWithConnection(sshAdv, task.Server, task.ExecuteCmd)
		}
	}
	output <- *taskForChannel(&task, strOutput, err, tasks.Finished, nil)
}

// 14.08.2022 20\:22\:06.816 ^.{0,10}[0-9]{1,2}[-\.][0-9]{1,2}[-\.]2[0-9]{3}[\s\t]+[0-9]{1,2}[\\:]{1,2}[0-9]{2}
// 15.08.2022 01:30:35.202
// 12.08.2022 02:14:35
// 2022-08-12	15:42:09 ^.{0,10}2[0-9]{3}[-\.][0-9]{1,2}[-\.][0-9]{1,2}[\s\t]+[0-9]{1,2}[\\:]{1,2}[0-9]{2}
// 2022.13.08 20\:46\:36.824 - same as above
// 2022.14.08 20:21:58.270 - same as above
// 2022.13.08 20\:46\:36.824 - same as above
// 2022.14.08 20:21:58.270	- same as above
// Aug 4, 2022 6:11:05 - ^.{0,10}[A-Za-z]{3} [0-9]{1,2}, 2[0-9]{3} [0-9]{1,2}:[0-9]{2}

//USE AWK instead of Grep:
// awk  '/Aug 4, 2022 6\:11\:05.*(Error|ERROR)/ {print $0; f = 1;
// if (match($0,/Aug 4, 2022/)) printf "MATCHED:[%s]", substr($0, RSTART, RLENGTH); next} f;/Aug 4, 2022/ {if (f == 1){f = 0; print "#######################"};}' ./admin.log

// AWK SIMPLE: awk  '/Aug 4, 2022 6\:11\:05.*(Error|ERROR)/{ print $0; f = 1; next } f; /Aug 4, 2022/ { if (f == 1){ f = 0; print "#######################"}}' ./admin.log
// Parse in GO ?

// AWK with seraching date
// # DATE:
//
