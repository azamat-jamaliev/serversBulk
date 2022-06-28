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
func ErrFatalWithMessage(msg string, err error) {
	color.Red("\n%s\n", msg)
	ErrFatal(err)
}
func ErrLog(err error) {
	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)
	boldRed.Printf("\nERROR: %s\n", err.Error())
}
func ErrLogWinMessage(msg string, err error) {
	color.Yellow("\n%s\n", msg)
	ErrLog(err)
}
func LogPrintln(message string) {
	LogPrintf("%s", message)
}
func LogPrintf(format string, v ...interface{}) {
	newFmrt := fmt.Sprintf("%s\n", format)
	fmt.Printf(newFmrt, v...)
}
