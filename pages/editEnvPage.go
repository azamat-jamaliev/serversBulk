package pages

import (
	"serversBulk/modules/configProvider"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func notEmpty(str string) bool {
	return len(strings.TrimSpace(str)) > 0
}
func addServer(editEnvForm *tview.Form, srv *configProvider.ConfigServerType) {
	editEnvForm.AddInputField("Server Name: ", srv.Name, 20, func(textToCheck string, lastChar rune) bool {
		return notEmpty(textToCheck)
	}, func(text string) {
		srv.Name = text
	}).
		AddTextArea("Log Folders:", strings.Join(srv.LogFolders[:], "\n"), 0, 3, 0, func(textToCheck string) {
			if notEmpty(textToCheck) {
				srv.LogFolders = strings.Split(textToCheck, "\n")
			}
		}).
		AddInputField("Log File Pattern: ", srv.LogFilePattern, 7, func(textToCheck string, lastChar rune) bool {
			return notEmpty(textToCheck)
		}, func(text string) {
			srv.LogFilePattern = text
		}).
		AddInputField("ssh login: ", srv.Login, 20, func(textToCheck string, lastChar rune) bool {
			return notEmpty(textToCheck)
		}, func(text string) {
			srv.Login = text
		}).
		AddPasswordField("ssh password: ", srv.Passowrd, 20, '*', func(text string) {
			srv.Passowrd = text
		}).
		AddInputField("ssh identity file: ", srv.IdentityFile, 60, nil, func(text string) {
			srv.IdentityFile = text
		}).
		AddTextArea("Servers IPs/Names:", strings.Join(srv.IpAddresses[:], "\n"), 0, 3, 0, func(textToCheck string) {
			if notEmpty(textToCheck) {
				srv.IpAddresses = strings.Split(textToCheck, "\n")
			}
		}).
		AddInputField(".    Bastion server (Jump server) IP/name: ", srv.BastionServer, 20, nil, func(text string) {
			srv.BastionServer = text
		}).
		AddInputField(".    Bastion ssh login: ", srv.BastionLogin, 20, nil, func(text string) {
			srv.BastionLogin = text
		}).
		AddPasswordField(".    Bastion ssh password: ", srv.BastionPassword, 20, '*', func(text string) {
			srv.BastionPassword = text
		}).
		AddInputField(".    Bastion Identity File: ", srv.BastionIdentityFile, 60, nil, func(text string) {
			srv.BastionIdentityFile = text
		})
}
func addBastion(editEnvForm *tview.Form, srv *configProvider.ConfigServerType) {

}

func EditEnvPage(appObj *tview.Application, env *configProvider.ConfigEnvironmentType, exitHandlerFunc func()) tview.Primitive {
	// var focusOrder []tview.Primitive
	editEnvForm := tview.NewForm()
	editEnvForm.SetBorder(true).SetTitle("Environment Information")

	modifyEnv := env
	// if env == nil {
	// 	modifyEnv = &configProvider.ConfigEnvironmentType{}
	// }
	editEnvForm.AddInputField("Environment name:", modifyEnv.Name, 20, nil, nil)
	for _, srv := range modifyEnv.Servers {
		addServer(editEnvForm, &srv)
	}

	page, _ := NewPageWithFooter(editEnvForm, "[Ctrl+Z]=Cancel&Exit [tab]=next field [Ctrl+A]=Add Servers group [Ctrl+S]=Save&Exit")

	appObj.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlZ {
			exitHandlerFunc()
		} else if event.Key() == tcell.KeyCtrlS {
			env = modifyEnv
			exitHandlerFunc()
		} else if event.Key() == tcell.KeyCtrlA {
			srvNew := configProvider.ConfigServerType{}
			modifyEnv.Servers = append(modifyEnv.Servers, srvNew)
			addServer(editEnvForm, &srvNew)
		}
		return event
	})

	return page
}
