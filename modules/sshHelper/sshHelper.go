package sshHelper

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"serversBulk/modules/configProvider"
	"serversBulk/modules/logHelper"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SshAdvanced struct {
	fileSftp       *sftp.File
	clientSftp     *sftp.Client
	client         *ssh.Client
	bastionSrvConn *net.Conn
	bastionClient  *ssh.Client
}

func (c *SshAdvanced) Close() {
	// if c.fileSftp != nil {
	// 	c.fileSftp.Close()
	// }
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
		logHelper.ErrFatal(err)
	}
	c.clientSftp = sc
	return c.clientSftp
}

// func (c *SshAdvanced) GetSftpFile(filePath string) *sftp.File {
// 	// Note: SFTP To Go doesn't support O_RDWR mode
// 	srcFile, err := c.newSftpClient().OpenFile(filePath, (os.O_RDONLY))
// 	if err != nil {
// 		logHelper.ErrFatalWithMessage(
// 			fmt.Sprintf("Unable to open file=[%s] on server=[%s]", filePath, c.client.RemoteAddr().String()),
// 			err)
// 	}
// 	c.fileSftp = srcFile
// 	return c.fileSftp
// }

func OpenSshAdvanced(serverConf *configProvider.ConfigServerType, server string) *SshAdvanced {
	logHelper.LogPrintf("OpenSshAdvanced connect to server:[%s]\n", server)
	sshAdvanced := SshAdvanced{}

	sshConfig := &ssh.ClientConfig{
		User: serverConf.Login,
		Auth: []ssh.AuthMethod{
			ssh.Password(serverConf.Passowrd),
		},
		Timeout: time.Minute * 5,
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	if serverConf.BastionServer != "" {
		var hostKey ssh.PublicKey
		// A public key may be used to authenticate against the remote
		// server by using an unencrypted PEM-encoded private key file.
		//
		// If you have an encrypted private key, the crypto/x509 package
		// can be used to decrypt it.
		key, err := ioutil.ReadFile(serverConf.BastionIdentityFile)
		if err != nil {
			logHelper.ErrFatalWithMessage("unable to read private key", err)
		}
		// Create the Signer for this private key.
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			logHelper.ErrFatalWithMessage("unable to parse private key", err)
		}

		sshConfigBastion := &ssh.ClientConfig{
			User: serverConf.BastionLogin,
			Auth: []ssh.AuthMethod{
				// Use the PublicKeys method for remote authentication.
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.FixedHostKey(hostKey),
			Timeout:         time.Minute * 5,
		}
		sshConfigBastion.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		// connect to the bastion host
		bastionConn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", serverConf.BastionServer), sshConfigBastion)
		if err != nil {
			logHelper.ErrFatalWithMessage(fmt.Sprintf("Cannot connect to Bastion server: %s", serverConf.BastionServer), err)
		}
		sshAdvanced.bastionClient = bastionConn

		// Dial a connection to the service host, from the bastion
		connBtoS, err := bastionConn.Dial("tcp", fmt.Sprintf("%s:22", server))
		if err != nil {
			log.Fatal(err)
		}
		sshAdvanced.bastionSrvConn = &connBtoS

		connCtoC, chans, reqs, err := ssh.NewClientConn(connBtoS, server, sshConfig)
		if err != nil {
			log.Fatal(err)
		}

		sshAdvanced.client = ssh.NewClient(connCtoC, chans, reqs)
		// sClient is an ssh client connected to the service host, through the bastion host.
	} else {
		logHelper.LogPrintf("connect to ssh server: %s", server)
		var e error
		sshAdvanced.client, e = ssh.Dial("tcp", fmt.Sprintf("%s:22", server), sshConfig)
		if e != nil {
			logHelper.ErrFatal(e)
		}
	}

	// // defer conn.Close()
	// if needSession {
	// 	session, _ := conn.NewSession()
	// 	// defer session.Close()
	// 	return conn, session
	// }
	return &sshAdvanced
}
