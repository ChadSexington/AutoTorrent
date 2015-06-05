package main

/* TODO
Finish up autotorrent.go
*/

import (
	"flag"
	"fmt"
	autotorrent "github.com/ChadSexington/AutoTorrent"
	"os"
)

const (
	actionUsage = "Perform a single action: erase, download, add, or complete."
	nameUsage   = "Name of a download to perform an action upon."
	idUsage     = "Local id of a download to perform an action upon."
	daemonUsage = "Run as a daemon, checking for finished torrents and downloading them."
	helpUsage   = "Show this usage output."
)

func main() {
	fmt.Println("Starting main")

	// Define command line arguments
	var action string
	var name string
	var id int
	var help bool
	var daemon bool
	flag.StringVar(&action, "action", "", actionUsage)
	flag.StringVar(&action, "a", "", actionUsage)
	flag.StringVar(&name, "name", "", nameUsage)
	flag.StringVar(&name, "n", "", nameUsage)
	flag.IntVar(&id, "id", 0, idUsage)
	flag.IntVar(&id, "i", 0, idUsage)
	flag.BoolVar(&help, "help", false, helpUsage)
	flag.BoolVar(&help, "h", false, helpUsage)
	flag.BoolVar(&daemon, "daemon", false, daemonUsage)
	flag.BoolVar(&daemon, "d", false, daemonUsage)

	// Parse defined options
	flag.Parse()

	switch {
	case flag.NFlag() == 0 || help:
		printUsage()
	case action != "" && (id == 0 && name == ""):
		fmt.Println("Must provide a name or id when performing an action.")
	case action != "" && !daemon:
		performAction(action, id, name)
	case daemon:
		if action != "" || name != "" || id != 0 {
			fmt.Println("All flags ignored when -d or --daemon is passed")
		}
		startDaemon()
	default:
		fmt.Println("Unrecognized options")
		printUsage()
	}

	fmt.Println("Ending main")
}

func startDaemon() {
	fmt.Println("Starting AutoTorrent daemon")
	at := autotorrent.New()
	err := at.StartDaemon()
	if err != nil {
		fmt.Println("Fatal daemon error: ", err.Error())
	}
}

func performAction(action string, id int, name string) {
	var byid bool
	var byname bool
	var err error

	if id != 0 {
		byid = true
	}
	if name != "" {
		byname = true
	}

	at := autotorrent.New()
	switch action {
	case "erase":
		if byid {
			err = at.EraseById(id)
		} else {
			err = at.EraseByName(name)
		}
	case "add":
		if byname {
			err = at.AddByName(name)
		} else {
			panic("Can only add by torrent name. ID will be automatically assigned.")
		}
	case "download":
		if byid {
			err = at.DownloadById(id)
		} else {
			err = at.DownloadByName(name)
		}
	case "complete":
		if byid {
			err = at.MarkCompleteById(id)
		} else {
			err = at.MarkCompleteByName(name)
		}
	default:
		fmt.Println("Action not recognized.")
		printUsage()
	}

	if err == nil {
		fmt.Println("Action completed successfully.")
	} else {
		fmt.Printf("Error occured while processed %s: %s.\n", action, err.Error())
	}
}

func printUsage() {
	fmt.Println("USAGE")
	os.Exit(1)
}
