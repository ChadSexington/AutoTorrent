package atsupport

// Provides functionality to read from a configuration file

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Conf struct {
	RemoteSSHUrl  string `yaml:"remote_ssh_url"`
	RemoteSSHUser string `yaml:"remote_ssh_user"`
	RemoteSSHKey  string `yaml:"remote_ssh_key"`

	TransmissionUrl      string `yaml:"transmission_url"`
	TransmissionUser     string `yaml:"transmission_user"`
	TransmissionPassword string `yaml:"transmission_password"`

	MysqlHost     string `yaml:"mysql_host"`
	MysqlPort     string `yaml:"mysql_port"`
	MysqlUser     string `yaml:"mysql_user"`
	MysqlPassword string `yaml:"mysql_password"`
	MysqlDatabase string `yaml:"mysql_database"`

	DownloadDir            string `yaml:"incomplete_download_dir"`
	CompletedDir           string `yaml:"completed_dowload_dir"`
	MaxConcurrentDownloads int    `yaml:"max_concurrent_downloads"`
}

func GetConfiguration() (c Conf) {
	confile, err := ioutil.ReadFile("/etc/autotorrent.yml")
	if err != nil {
		panic("Failed to read configuration: " + err.Error())
	}
	err = yaml.Unmarshal([]byte(confile), &c)
	if err != nil {
		panic("Failed to read configuration: " + err.Error())
	}
	return c
}
