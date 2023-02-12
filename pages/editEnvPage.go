package pages

import (
	"serversBulk/modules/configProvider"
	"strings"

	"github.com/rivo/tview"
)

func EditEnvPage(appObj *tview.Application, env *configProvider.ConfigEnvironmentType, exitHandlerFunc func()) tview.Primitive {
	// var focusOrder []tview.Primitive
	editEnvForm := tview.NewForm()
	editEnvForm.SetBorder(true).SetTitle("Environment Information")

	modifyEnv := env
	if env == nil {
		modifyEnv = &configProvider.ConfigEnvironmentType{}
	}
	editEnvForm.AddInputField("Environment name:", modifyEnv.Name, 20, nil, nil)
	for _, srv := range modifyEnv.Servers {
		editEnvForm.AddInputField("Server Name: ", srv.Name, 20, nil, nil).
			AddTextArea("Log Folders:", strings.Join(srv.LogFolders[:], "\n"), 0, 3, 0, nil).
			AddInputField("Log File Pattern: ", srv.LogFilePattern, 7, nil, nil).
			AddInputField("ssh login: ", srv.Login, 20, nil, nil).
			AddPasswordField("ssh password: ", srv.Passowrd, 20, '*', nil).
			AddInputField("ssh identity file: ", srv.IdentityFile, 60, nil, nil).
			AddTextArea("Servers IPs/Names:", strings.Join(srv.IpAddresses[:], "\n"), 0, 3, 0, nil).
			AddCheckbox("Use Bastion (ssh tunnel):", srv.BastionServer != "", nil)
		if srv.BastionServer != "" {
			editEnvForm.AddInputField("Bastion server IP/name: ", srv.BastionServer, 20, nil, nil).
				AddInputField("Bastion ssh login: ", srv.BastionLogin, 20, nil, nil).
				AddPasswordField("Bastion ssh password: ", srv.BastionPassword, 20, '*', nil).
				AddInputField("Bastion Identity File: ", srv.BastionIdentityFile, 60, nil, nil)
		}
	}
	editEnvForm.AddButton(" -      Add Servers        - ", nil).
		AddButton("Save", nil).
		AddButton("Cancel", func() {
			exitHandlerFunc()
		})

	return editEnvForm
}
