package autotorrent

import (
	"fmt"
	atsupport "github.com/ChadSexington/AutoTorrent/lib"
	"github.com/ChadSexington/go-transmission/transmission"
	"path/filepath"
	"sync"
	"time"
)

type AutoTorrent struct {
	Transmission transmission.TransmissionClient
	Conf         atsupport.Conf
	Datastore    atsupport.Datastore
}

func New() AutoTorrent {
	conf := atsupport.GetConfiguration()
	trans := transmission.New(conf.TransmissionUrl, conf.TransmissionUser, conf.TransmissionPassword)
	ds, err := atsupport.NewDatastore(conf.MysqlHost, conf.MysqlPort, conf.MysqlUser, conf.MysqlPassword, conf.MysqlDatabase)
	if err != nil {
		panic(err.Error())
	}
	return AutoTorrent{Transmission: trans, Conf: conf, Datastore: ds}
}

// Start the daemon
func (at *AutoTorrent) StartDaemon() error {
	for {
		// Get list of torrents from transmissoin
		torrentList, err := at.Transmission.GetDownloadedTorrents()
		if err != nil {
			fmt.Println("Unable to get completed torrent list with error", err)
			return err
		}

		// Instantiate a waitgroup
		var wg sync.WaitGroup

		currentDownloads := 0
		var preExisting bool
		// Loop through torrents
		// to download the torrents
		for x := 0; x < len(torrentList); x++ {
			torrent := torrentList[x]
			preExisting = false

			// Create download in database for new torrent
			dl, err := at.Datastore.NewDownload(torrent.Name, torrent.DownloadDir, false)
			if err != nil {
				// Handle pre-existing entries
				if err.Error() == "Duplicate entry" {
					if dl.Complete {
						fmt.Printf("Torrent %s already downloaded, skipping...\n", torrent.Name)
						continue
					} else {
						fmt.Printf("Torrent with name %s already in database, but not finished. Downloading...\n", torrent.Name)
						preExisting = true
						dl, err = at.Datastore.GetDownloadByName(torrent.Name)
						if err != nil {
							fmt.Printf("Could not get download for torrent %s, skipping...\n", torrent.Name)
							continue
						}
					}
				} else {
					fmt.Printf("Error getting download for torrent %s: %s\n", torrent.Name, err)
					fmt.Printf("Skipping download for torrent %s due to above errors\n", torrent.Name)
					continue
				}
			}

			if !preExisting {
				// Create download file entries
				for x := range torrent.Files {
					torFile := torrent.Files[x]
					remotePath := filepath.Join(torrent.DownloadDir, torFile.Name)
					localPath := filepath.Join(at.Conf.DownloadDir, torFile.Name)
					dlFile, err := at.Datastore.NewDownloadFile(torFile.Name, remotePath, localPath, false, dl.ID)
					if err != nil {
						return err
					}
					// Add entry to download
					dl.Files = append(dl.Files, dlFile)
				}
			}

			fmt.Printf("Created or found existing download for %s\n", torrent.Name)

			// If max downloads are happening, wait
			for currentDownloads >= at.Conf.MaxConcurrentDownloads {
				time.Sleep(time.Second * 30)
			}

			wg.Add(1)
			// Download the torrent in a goroutine
			go func() {
				currentDownloads = currentDownloads + 1
				defer func() { currentDownloads = currentDownloads - 1 }()
				defer wg.Done()
				fmt.Println("Downloading torrent", torrent.Name)
				err = atsupport.DownloadTorrent(torrent, dl, at.Datastore, at.Conf)
				if err != nil {
					fmt.Printf("Unable to download torrent %s with error: %s\n", torrent.Name, err)
				}
				fmt.Println("Download of torrent Complete:", torrent.Name)
				err = at.Datastore.DownloadComplete(dl)
				if err != nil {
					fmt.Printf("Unable to mark download of %s complete with error: %s\n", torrent.Name, err)
				}
			}()
		} //end for

		if currentDownloads == 0 {
			fmt.Println("No new downloads, waiting a bit and looking again")
			time.Sleep(time.Second * 30)
		} else {
			fmt.Printf("All downloads started, waiting for completion.")
			// Wait for downloads to complete
			wg.Wait()
			fmt.Println("All downloads completed")
		}
	}
	return nil
}

func (at *AutoTorrent) DownloadByName(name string) error {
	// Need to devise a way to get a single torrent by name from transmission
	// Probably need to make an edit at go-transmission
	/*
		dl, err := at.Datastore.GetDownloadByName(name)
		if err != nil {
			fmt.Println("Error reading database with error", err)
			return err
		}
		err = atsupport.DownloadTorrent(torrent, dl)
		if err != nil {
			fmt.Printf("Unable to download torrent %s with error: %s\n", torrent.Name, err)
		}
		err = at.Datastore.DownloadComplete(dl)
		if err != nil {
			fmt.Printf("Unable to mark dowload of %s complete with error: %s\n", torrent.Name, err)
		}
	*/
	return nil
}

func (at *AutoTorrent) DownloadById(id int) error {
	// Need to devise a way to get a single torrent by name from transmission
	// Probably need to make an edit at go-transmission
	return nil
}

func (at *AutoTorrent) EraseByName(name string) error {
	// Does this erase from the local database or the server?
	// We need a method for both.
	return nil
}

func (at *AutoTorrent) EraseById(id int) error {
	return nil
}

func (at *AutoTorrent) AddByName(name string) error {
	return nil
}

func (at *AutoTorrent) MarkCompleteByName(name string) error {
	return nil
}

func (at *AutoTorrent) MarkCompleteById(id int) error {
	return nil
}
