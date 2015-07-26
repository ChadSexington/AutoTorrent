package atsupport

// Misc functionality

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ChadSexington/go-transmission/transmission"
)

// Download all files from a torrent on the remote server
// Will return nil if a download has already been downloaded.
func DownloadTorrent(torrent transmission.Torrent, dl Download, ds Datastore, conf Conf) (err error) {
	// Create ssh session
	conn, err := getSSHSession(conf.RemoteSSHUser, conf.RemoteSSHKey, conf.RemoteSSHUrl)
	defer conn.Close()

	var dlFile DownloadFile
	// Loop through each file
	for x := range dl.Files {
		dlFile = dl.Files[x]
		fmt.Println("Downloading: ", dl.Files[x].Name)
		// Return if file already downloaded.
		if dlFile.Complete {
			fmt.Printf("Not downloading %s, file already exists.\n", dlFile.Name)
			continue
		}
		// Create directories if they don't exist
		err := createDir(filepath.Dir(dlFile.LocalPath))
		if err != nil {
			return err
		}

		// Download the file
		err = downloadFile(dlFile, conn)
		if err != nil {
			return err
		}
		ds.DownloadFileComplete(dlFile)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Files for torrent %s complete. Moving to completed directory...\n", torrent.Name)

	// When all files are done downloading, move to completedDir
	// Going to just have to go through each file, createDir on each, then copy the file.
	for x := range dl.Files {
		dlFile := dl.Files[x]
		newLocation := strings.Replace(dlFile.LocalPath, conf.DownloadDir, conf.CompletedDir, 1)
		err := createDir(filepath.Dir(newLocation))
		if err != nil {
			return err
		}
		ds.UpdateDownloadFileLocalPath(dlFile, newLocation)
		if err != nil {
			return err
		}

		fmt.Printf("Moving %s to %s...\n", dlFile.LocalPath, newLocation)
		err = os.Rename(dlFile.LocalPath, newLocation)
		if err != nil {
			return err
		}
		fmt.Printf("Move complete to %s\n", newLocation)
	}

	return nil
}

func getSSHSession(remoteSSHUser, remoteSSHKey, remoteSSHUrl string) (conn *ssh.Client, err error) {
	// Set up authentication
	privateKey, err := ioutil.ReadFile(remoteSSHKey)
	if err != nil {
		return conn, err
	}
	signer, _ := ssh.ParsePrivateKey([]byte(privateKey))
	clientConfig := &ssh.ClientConfig{
		User: remoteSSHUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	// Create ssh connection
	fmt.Println("Creating ssh connection")
	var sshCreated bool
	for i := 0; i < 5; i++ {
		conn, err = ssh.Dial("tcp", remoteSSHUrl, clientConfig)
		if err != nil {
			fmt.Println("Failed to dial: " + err.Error())
			sshCreated = false
		} else {
			sshCreated = true
			break
		}
	}

	if !sshCreated {
		panic("Unable to start ssh connection: " + err.Error())
	}

	return conn, err

}

// Create all directories in a path if they do not already exist
// returns nil if creation succeeded, or if directories already exist
func createDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		mkdirErr := os.MkdirAll(dir, 0755)
		if mkdirErr != nil {
			return mkdirErr
		}
	}
	return nil
}

// Download a file from remotePath to localPath
func downloadFile(dlFile DownloadFile, conn *ssh.Client) error {
	// Create sftp client
	sftp, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer sftp.Close()

	// Open file for I/O
	remoteFile, err := sftp.Open(dlFile.RemotePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()
	localFile, err := os.Create(dlFile.LocalPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	// Transfer file
	fmt.Printf("Transferring remote file %s to local %s\n", dlFile.RemotePath, dlFile.LocalPath)
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return err
	}

	// Set permissions
	err = os.Chmod(dlFile.LocalPath, 0755)
	if err != nil {
		fmt.Printf("Unable to set permissions for %s, continuing anyway.\n", dlFile.LocalPath)
		fmt.Printf("Error: %s", err.Error())
	}

	fmt.Println("Download of file complete: ", dlFile.Name)

	return nil
}
