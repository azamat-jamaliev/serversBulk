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

func getSshSigner(identityFilePath string) ssh.Signer {
	// serverConf.BastionIdentityFile
	// A public key may be used to authenticate against the remote
	// server by using an unencrypted PEM-encoded private key file.
	//
	// If you have an encrypted private key, the crypto/x509 package
	// can be used to decrypt it.
	key, err := ioutil.ReadFile(identityFilePath)
	if err != nil {
		panic(err)
	}
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		panic(err)
	}

	return signer
}

func OpenSshAdvanced(serverConf *configProvider.ConfigServerType, server string) *SshAdvanced {
	sshAdvanced := SshAdvanced{}

	var srvAuthMethod []ssh.AuthMethod
	if serverConf.IdentityFile != "" {
		srvAuthMethod = []ssh.AuthMethod{
			ssh.PublicKeys(getSshSigner(serverConf.IdentityFile)),
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

	if serverConf.BastionServer != "" {

		var hostKey ssh.PublicKey
		var authMethod []ssh.AuthMethod

		if serverConf.BastionIdentityFile != "" {
			authMethod = []ssh.AuthMethod{
				ssh.PublicKeys(getSshSigner(serverConf.BastionIdentityFile)),
			}
		} else {
			authMethod = []ssh.AuthMethod{
				ssh.Password(serverConf.BastionPassword),
			}
		}

		sshConfigBastion := &ssh.ClientConfig{
			User:            serverConf.BastionLogin,
			Auth:            authMethod,
			HostKeyCallback: ssh.FixedHostKey(hostKey),
			Timeout:         time.Minute * 5,
		}
		sshConfigBastion.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		// connect to the bastion host
		bastionConn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", serverConf.BastionServer), sshConfigBastion)
		if err != nil {
			panic(fmt.Sprintf("Cannot connect to Bastion server: %s\nERROR:%v", serverConf.BastionServer, err))
		}
		sshAdvanced.bastionClient = bastionConn

		// Dial a connection to the service host, from the bastion
		connBtoS, err := bastionConn.Dial("tcp", fmt.Sprintf("%s:22", server))
		if err != nil {
			panic(err)
		}
		sshAdvanced.bastionSrvConn = &connBtoS

		connCtoC, chans, reqs, err := ssh.NewClientConn(connBtoS, server, sshConfig)
		if err != nil {
			panic(err)
		}

		sshAdvanced.client = ssh.NewClient(connCtoC, chans, reqs)
		// sClient is an ssh client connected to the service host, through the bastion host.
	} else {
		var e error
		sshAdvanced.client, e = ssh.Dial("tcp", fmt.Sprintf("%s:22", server), sshConfig)
		if e != nil {
			panic(e)
		}
	}

	return &sshAdvanced
}
