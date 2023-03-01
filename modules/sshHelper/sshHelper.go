package sshHelper

import (
	"fmt"
	"io/ioutil"
	"net"
	"sebulk/modules/configProvider"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SshAdvanced struct {
	clientSftp     *sftp.Client
	client         *ssh.Client
	bastionSrvConn *net.Conn
	bastionClient  *ssh.Client
}

func (c *SshAdvanced) Close() {
	if c.clientSftp != nil {
		c.clientSftp.Close()
	}
	if c.client != nil {
		c.client.Close()
	}
	if c.bastionSrvConn != nil {
		(*c.bastionSrvConn).Close()
	}
	if c.bastionClient != nil {
		c.bastionClient.Close()
	}
}
func (c *SshAdvanced) NewSession() *ssh.Session {
	sess, _ := c.client.NewSession()
	return sess
}
func (c *SshAdvanced) NewSftpClient() *sftp.Client {
	// Create new SFTP client
	sc, err := sftp.NewClient(c.client)
	if err != nil {
		panic(err)
	}
	c.clientSftp = sc
	return c.clientSftp
}

func getSshSigner(identityFilePath string) (ssh.Signer, error) {
	var signer ssh.Signer
	key, err := ioutil.ReadFile(identityFilePath)
	if err == nil {
		signer, err = ssh.ParsePrivateKey(key)
	}
	return signer, err
}

func OpenSshAdvanced(serverConf *configProvider.ConfigServerType, server string) (*SshAdvanced, error) {
	var signer ssh.Signer
	var err error

	sshAdvanced := SshAdvanced{}

	var srvAuthMethod []ssh.AuthMethod
	if serverConf.IdentityFile != "" {
		signer, err = getSshSigner(serverConf.IdentityFile)
		srvAuthMethod = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else {
		srvAuthMethod = []ssh.AuthMethod{
			ssh.Password(serverConf.Passowrd),
		}
	}

	sshConfig := &ssh.ClientConfig{
		User:    serverConf.Login,
		Auth:    srvAuthMethod,
		Timeout: time.Minute * 5,
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	if err == nil {
		if serverConf.BastionServer != "" {

			var hostKey ssh.PublicKey
			var authMethod []ssh.AuthMethod

			if serverConf.BastionIdentityFile != "" {
				signer, err = getSshSigner(serverConf.BastionIdentityFile)
				authMethod = []ssh.AuthMethod{
					ssh.PublicKeys(signer),
				}
			} else {
				authMethod = []ssh.AuthMethod{
					ssh.Password(serverConf.BastionPassword),
				}
			}

			if err == nil {
				sshConfigBastion := &ssh.ClientConfig{
					User:            serverConf.BastionLogin,
					Auth:            authMethod,
					HostKeyCallback: ssh.FixedHostKey(hostKey),
					Timeout:         time.Minute * 5,
				}
				sshConfigBastion.HostKeyCallback = ssh.InsecureIgnoreHostKey()
				// connect to the bastion host
				bastionConn, e := ssh.Dial("tcp", fmt.Sprintf("%s:22", serverConf.BastionServer), sshConfigBastion)
				if e != nil {
					return nil, fmt.Errorf("CANNOT CONNECT TO BASTION SERVER: %s\nERROR:%v", serverConf.BastionServer, e)
				}
				sshAdvanced.bastionClient = bastionConn

				// Dial a connection to the service host, from the bastion
				connBtoS, e := bastionConn.Dial("tcp", fmt.Sprintf("%s:22", server))
				if e != nil {
					return nil, fmt.Errorf("CANNOT CONNECT TO SERVER: %s VIA BASTION SERVER:%s \nERROR:%v", server, serverConf.BastionServer, e)
				}
				sshAdvanced.bastionSrvConn = &connBtoS

				connCtoC, chans, reqs, e := ssh.NewClientConn(connBtoS, server, sshConfig)
				if e != nil {
					return nil, fmt.Errorf("CANNOT CREATE CONNECTION TO SERVER: %s BASTION SERVER:%s \nERROR:%v", server, serverConf.BastionServer, e)
				}

				sshAdvanced.client = ssh.NewClient(connCtoC, chans, reqs)
				// sClient is an ssh client connected to the service host, through the bastion host.
			}
		} else {
			sshAdvanced.client, err = ssh.Dial("tcp", fmt.Sprintf("%s:22", server), sshConfig)
		}
	}

	return &sshAdvanced, err
}
