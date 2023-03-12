package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"sebulk/modules/sshHelper"
	"sebulk/modules/tasks"
)

var dateRegExp = []string{
	"[0-9]{2}[-\\.][0-9]{1,2}[-\\.]2[0-9]{3} [0-9]{1,2}[\\:]{1,2}[0-9]{2}[\\:]{1,2}[0-9]{2}",
	"2[0-9]{3}[-\\.][0-9]{1,2}[-\\.][0-9]{1,2} [0-9]{1,2}[\\:]{1,2}[0-9]{2}[\\:]{1,2}[0-9]{2}",
	"[A-Za-z]{3} [0-9]{1,2}, 2[0-9]{3} [0-9]{1,2}:[0-9]{2}:[0-9]{2} [PAM]{2}",
}

const charsBeforeDate = "^.{0,20}"

var datesTemplas = []string{
	"02.01.2006 15\\:04\\:05",
	"02.01.2006 15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05",
	"2006.02.01 15\\:04\\:05",
	"2006.02.01 15:04:05",
	"Jan _2, 2006 3:04:05 PM",
}
var strRegExDates = strings.Join(dateRegExp, "|")

// NOTE: output result is in the channel
func awkInLogs(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	// TODO: improve serach by showing only data for specified period:
	statusHandler(task.Server, "CONNECTING...")
	logHandler(task.Server, fmt.Sprintf("connecting to server: [%s] to awk", task.Server))
	strOutput := ""
	sshAdv, err := sshHelper.OpenSshAdvanced(&task.ConfigServer, task.Server)
	var curSrvTime time.Time
	if err == nil {
		defer sshAdv.Close()
		strOutput, err = executeWithConnection(sshAdv, task.Server, `date -u +"%Y-%m-%dT%H:%M:%S"`)
		//
		if err == nil {
			strOutput = strings.TrimSpace(strOutput)
			curSrvTime, err = time.Parse("2006-01-02T15:04:05", strOutput)
			if err == nil {
				mtime := 0.01
				mtime, err = strconv.ParseFloat(task.ModifTime, 32)
				if err == nil && mtime < 0 {
					hours := math.Round((mtime * 24) + 0.5)
					fromTime := curSrvTime.Add(time.Duration(hours) * time.Hour)
					// ^.{0,10}[0-9]{1,2}[-\.][0-9]{1,2}[-\.]2[0-9]{3} [0-9]{1,2}[\\:]{1,2}[0-9]{2}

					strAwk := fmt.Sprintf("awk  '/%s(%s)[^\\n]{0,60}(Error|ERROR)/{ print $0; f = 1 ;next } f; /(%s)/ { if (f == 1){ f = 0; print \"+++++++++++\"}}' {} ", charsBeforeDate, strRegExDates, strRegExDates)
					task.ExecuteCmd = getFindExecForTask(task, strAwk)
					strOutput, err = executeWithConnection(sshAdv, task.Server, task.ExecuteCmd)

					if err == nil {
						strOutput = filterLogLines(strOutput, fromTime)
					}
				}
			}
		}
	}
	output <- *taskForChannel(&task, strOutput, err, tasks.Finished, nil)
}
func filterLogLines(logTest string, timeFrom time.Time) string {
	logDatRegex := fmt.Sprintf("%s(%s)", charsBeforeDate, strRegExDates)
	re := regexp.MustCompile(logDatRegex)
	logLines := strings.Split(logTest, "\n")

	begin := false
	chunk := ""
	strOutput := ""
	var err error
	for _, line := range logLines {
		matchRe := re.FindStringSubmatch(line)
		if len(matchRe) > 1 {
			if len(chunk) > 0 {
				strOutput = fmt.Sprintf("%s\n%s", strOutput, chunk)
			}
			dateInLog := time.Now()
			for _, tmpl := range datesTemplas {
				dateInLog, err = time.Parse(tmpl, matchRe[1])
				if err == nil { //converted succesfully
					break
				}
			}
			if err != nil {
				panic(fmt.Sprintf("cannot convert value[%s] to date (from line:[%s])", matchRe[1], line))
			}
			begin = dateInLog.After(timeFrom)
		}
		if begin {
			chunk = fmt.Sprintf("%s\n%s", chunk, line)
		}
	}
	if len(chunk) > 0 {
		strOutput = fmt.Sprintf("%s\n%s", strOutput, chunk)
	}
	return strOutput
}
