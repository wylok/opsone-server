package common

import (
	"golang.org/x/crypto/ssh"
	"time"
)

func SshClient(host, port, username, passwd, pkey string) (*ssh.Client, error) {
	var auth []ssh.AuthMethod
	var err error
	if pkey != "" && pkey != "none" {
		signer, e := ssh.ParsePrivateKey([]byte(pkey))
		if e == nil {
			auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
		} else {
			err = e
		}
	} else {
		if passwd != "" && passwd != "none" {
			auth = []ssh.AuthMethod{ssh.Password(passwd)}
		}
	}
	config := &ssh.ClientConfig{
		User:            username,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	client, e := ssh.Dial("tcp", host+":"+port, config)
	if e != nil {
		err = e
	}
	return client, err
}
