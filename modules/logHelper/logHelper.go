package logHelper

import (
	"fmt"

	"github.com/fatih/color"
)

func ErrFatal(err error) {
	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)
	redStr := boldRed.Sprintf("\n!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\nFATAL ERROR: %s\n!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n", err.Error())
	panic(redStr)
}
func ErrFatalln(err error, msg string) {
	ErrFatalf(err, "%s", msg)
}
func ErrFatalf(err error, format string, v ...interface{}) {
	newFmrt := fmt.Sprintf(format, v...)
	color.Red("\n%s\n", newFmrt)
	ErrFatal(err)
}
func ErrLog(err error) {
	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)
	boldRed.Printf("\nERROR: %s\n", err.Error())
}
func ErrLogln(err error, msg string) {
	ErrLogf(err, "%s", msg)
}
func ErrLogf(err error, format string, v ...interface{}) {
	newFmrt := fmt.Sprintf(format, v...)
	color.Yellow("\n%s\n", newFmrt)
	ErrLog(err)
}
func LogPrintln(message string) {
	LogPrintf("%s\n", message)
}
func LogPrintf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
	fmt.Println("")
}
