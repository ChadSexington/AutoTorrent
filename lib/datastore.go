package atsupport

// Funtionality for interacting with the database

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)

// db schema
// downloads
//  - id (auto, int)
//  - name (CHAR50)
//  - finished (BOOL)
// files
//  - id (auto, int)
//  - name (CHAR100)
//  - finished (BOOL)
//	- remote_path (CHAR150)
//	- local_path (CHAR150
//  - download_id (int)

// db tut
// http://go-database-sql.org

type Datastore struct {
	host       string
	port       string
	username   string
	password   string
	database   string
	connection *sql.DB
}

type Download struct {
	ID       int
	Name     string
	Files    []DownloadFile
	Complete bool
}

type DownloadFile struct {
	ID         int
	Name       string
	Size       int
	RemotePath string
	LocalPath  string
	Complete   bool
}

// Get Datastore object
func NewDatastore(host, port, username, password, database string) (ds Datastore, err error) {
	conf := GetConfiguration()
	connStr := conf.MysqlUser + ":" + conf.MysqlPassword + "@tcp(" + conf.MysqlHost + ":" + conf.MysqlPort + ")/" + conf.MysqlDatabase
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return ds, err
	} else {
		err = db.Ping()
		if err != nil {
			return ds, err
		}
	}
	ds = Datastore{host: host, port: port, username: username, password: password, database: database, connection: db}
	return ds, nil
}

// Create a new download
func (ds *Datastore) NewDownload(name, torrentDownloadDir string, complete bool) (dl Download, err error) {
	// Create download entry
	stmt, err := ds.connection.Prepare("INSERT INTO downloads(id, name, finished) VALUES (?, ?, ?)")
	if err != nil {
		return dl, err
	}
	defer stmt.Close()
	var cmp string
	if complete {
		cmp = "1"
	} else {
		cmp = "0"
	}
	res, err := stmt.Exec("NULL", name, cmp)
	if err != nil {
		if strings.Contains(err.Error(), "Error 1062: Duplicate entry") {
			dl, err := ds.GetDownloadByName(name)
			if err != nil {
				return dl, err
			}
			return dl, errors.New("Duplicate entry")
		} else {
			return dl, err
		}
	}
	id, err := res.LastInsertId()
	if err != nil {
		// should remove row if this errors
		return dl, errors.New("Download was not created.")
	}

	dl = Download{ID: int(id), Name: name, Complete: complete, Files: []DownloadFile{}}
	return dl, nil
}

// Create download file entry
func (ds *Datastore) NewDownloadFile(name, remotePath, localPath string, complete bool, downloadId int) (dlFile DownloadFile, err error) {
	stmt, err := ds.connection.Prepare("INSERT INTO files(id, name, finished, remote_path, local_path, download_id) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return dlFile, err
	}
	defer stmt.Close()
	res, err := stmt.Exec("NULL", name, complete, remotePath, localPath, downloadId)
	if err != nil {
		// We don't care so much about duplicate file name entries.
		if !strings.Contains(err.Error(), "Error 1062: Duplicate entry") {
			return dlFile, err
		}
	}
	id, err := res.LastInsertId()
	if err != nil {
		// should remove row if this errors
		return dlFile, errors.New("Download file was not created.")
	}
	dlFile = DownloadFile{ID: int(id), Name: name, RemotePath: remotePath, LocalPath: localPath}
	return dlFile, nil
}

// Destroy a download by Id
func (ds *Datastore) DestroyDownloadById(id int) (err error) {
	// Destroy the files
	stmt, err := ds.connection.Prepare("DELETE FROM files WHERE download_id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	// Destroy the download
	stmt, err = ds.connection.Prepare("DELETE FROM downloads WHERE id=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

// Destroy a download by Name
func (ds *Datastore) DestroyDownloadByName(name string) (err error) {
	// Get the download
	download, err := ds.GetDownloadByName(name)
	if err != nil {
		return err
	}
	// Destroy the files
	stmt, err := ds.connection.Prepare("DELETE FROM files WHERE download_id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(download.ID)
	if err != nil {
		return err
	}
	// Destroy the download
	stmt, err = ds.connection.Prepare("DELETE FROM downloads WHERE name=\"?\"")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(name)
	if err != nil {
		return err
	}
	return nil
}

// Get a download by Id
func (ds *Datastore) GetDownloadById(id int) (dl Download, err error) {
	row := ds.connection.QueryRow("SELECT name,finished FROM downloads WHERE id=?", id)
	var name string
	var finished bool
	var files []DownloadFile
	err = row.Scan(&name, &finished, &files)
	if err != nil {
		if !strings.Contains(err.Error(), "no rows in result set") {
			return dl, err
		}
	}

	dl = Download{ID: id, Name: name, Complete: finished, Files: files}
	return dl, nil
}

// Get a download by Name
func (ds *Datastore) GetDownloadByName(name string) (dl Download, err error) {
	var id int
	var finished bool

	stmt, err := ds.connection.Prepare("SELECT id,finished FROM downloads WHERE name = ?")
	if err != nil {
		return dl, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(name).Scan(&id, &finished)
	if err != nil {
		return dl, err
	}

	dl = Download{ID: id, Name: name, Complete: finished, Files: []DownloadFile{}}
	files, err := ds.GetDownloadFiles(dl)
	if err != nil {
		return dl, err
	}
	dl.Files = files
	return dl, nil
}

// Get downloadFile objects for a specific download
func (ds *Datastore) GetDownloadFiles(dl Download) (files []DownloadFile, err error) {
	rows, err := ds.connection.Query("SELECT * FROM files WHERE download_id=?", dl.ID)
	if err != nil {
		return files, err
	}
	defer rows.Close()
	var id int
	var name string
	var finished bool
	var remotePath string
	var localPath string
	var downloadId int
	var dlFile DownloadFile
	for rows.Next() {
		err = rows.Scan(&id, &name, &finished, &remotePath, &localPath, &downloadId)
		if err != nil {
			return files, err
		}
		dlFile = DownloadFile{ID: id, Name: name, Complete: finished, RemotePath: remotePath, LocalPath: localPath}
		files = append(files, dlFile)
	}
	return files, nil
}

// Set a download complete
func (ds *Datastore) DownloadComplete(dl Download) error {
	for dlFile := range dl.Files {
		stmt, err := ds.connection.Prepare("UPDATE files SET finished=1 WHERE id=?")
		if err != nil {
			return err
		}

		_, err = stmt.Exec(dl.Files[dlFile].ID)
		if err != nil {
			return err
		}
		stmt.Close()
	}
	stmt, err := ds.connection.Prepare("UPDATE downloads SET finished=1 WHERE id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(dl.ID)
	if err != nil {
		return err
	}

	return nil
}

// Set a file complete
func (ds *Datastore) DownloadFileComplete(dlFile DownloadFile) error {
	stmt, err := ds.connection.Prepare("UPDATE files SET finished=true WHERE id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(dlFile.ID)
	if err != nil {
		return err
	}

	return nil
}

func (ds *Datastore) UpdateDownloadFileLocalPath(dlFile DownloadFile, path string) error {
	stmt, err := ds.connection.Prepare("UPDATE files SET path=? WHERE id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(path, dlFile.ID)
	if err != nil {
		return err
	}

	return nil
}
