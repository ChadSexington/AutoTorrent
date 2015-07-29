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
	var currentDownloads []int
MainLoop:
	for {
		// Get list of torrents from transmissoin
		torrentList, err := at.Transmission.GetDownloadedTorrents()
		if err != nil {
			fmt.Println("Unable to get completed torrent list with error", err)
			fmt.Println("Waiting 30 seconds and trying again...")
			time.Sleep(30 * time.Second)
			continue MainLoop
		}

		// Instantiate a waitgroup
		var wg sync.WaitGroup

		var preExisting bool
		// Loop through torrents
		// to download the torrents
	TorrentLoop:
		for x := 0; x < len(torrentList); x++ {
			torrent := torrentList[x]
			preExisting = false

			// Create download in database for new torrent
			dl, err := at.Datastore.NewDownload(torrent.Name, torrent.DownloadDir, false)
			if err != nil {
				// Handle pre-existing entries
				if err.Error() == "Duplicate entry" {
					if dl.Complete {
						// Skip download
						continue TorrentLoop
					} else {
						fmt.Printf("Torrent with name %s already in database, but not finished. Downloading...\n", torrent.Name)
						preExisting = true
						dl, err = at.Datastore.GetDownloadByName(torrent.Name)
						if err != nil {
							fmt.Printf("Could not get download for torrent %s, skipping...\n", torrent.Name)
							continue TorrentLoop
						}
					}
				} else {
					fmt.Printf("Error getting download for torrent %s: %s\n", torrent.Name, err)
					fmt.Printf("Skipping download for torrent %s due to above errors\n", torrent.Name)
					continue TorrentLoop
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
						fmt.Println("Unable to create new download file for new torrent with error: ", err)
					} else {
						// Add entry to download
						dl.Files = append(dl.Files, dlFile)
					}
				}
			}

			fmt.Printf("Created or found existing download for %s\n", torrent.Name)

			// If max downloads are happening, wait
			for len(currentDownloads) >= at.Conf.MaxConcurrentDownloads {
				time.Sleep(time.Second * 30)
			}

			for _, x := range currentDownloads {
				if x == dl.ID {
					fmt.Printf("%s is already downloading, skipping...\n", dl.Name)
					continue TorrentLoop
				}
			}

			wg.Add(1)
			// Download the torrent in a goroutine
			go func() {
				currentDownloads = append(currentDownloads, dl.ID)
				// Ensure download id is removed from slice when download is complete
				defer func() {
					for i, x := range currentDownloads {
						if x == dl.ID {
							// Remove from currentDownloads slice
							// https://github.com/golang/go/wiki/SliceTricks
							currentDownloads = append(currentDownloads[:i], currentDownloads[i+1:]...)
						}
					}
				}()
				// Ensure waitgroup is cleared after goroutine finishes.
				defer wg.Done()
				fmt.Println("Downloading torrent", torrent.Name)
				err = atsupport.DownloadTorrent(torrent, dl, at.Datastore, at.Conf)
				if err != nil {
					fmt.Printf("Unable to download torrent %s with error: %s\n", torrent.Name, err)
				} else {
					fmt.Println("Download of torrent Complete:", torrent.Name)
					err = at.Datastore.DownloadComplete(dl)
					if err != nil {
						fmt.Printf("Unable to mark download of %s complete with error: %s\n", torrent.Name, err)
					}
				}
			}()
		} //end torrent list loop

		// Wait between each time contacting the torrent server
		time.Sleep(300 * time.Second)
		/*
			if currentDownloads == 0 {
				fmt.Println("No new downloads, waiting a bit and looking again")
				time.Sleep(time.Second * 30)
			} else {
				fmt.Printf("All downloads started, waiting for completion.")
				// Wait for downloads to complete
				wg.Wait()
				fmt.Println("All downloads completed")
			}
		*/
	} //end main loop
	fmt.Println("Somehow got out of main loop, ending...")
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
