#AutoTorrent

This command line tool can automate secure downloads from a remote torrenting server running Transmission to a local directory.

##Requirements:

- Mysql server
- Remote Transmission server with SSH access

##Setup:

1) Configure a mysql database with the follow table(s):
		- downloads
			- id (int), name (string), finished (bool)
2) Create and configure an ssh key to the remote server
3) Copy the autotorrent.yml.example to /etc/ and rename to /etc/autotorrent.yml
4) Fill out the configuration file.
5) Run the tool as a daemon with:
```
$ autotorrent_cli -d
```

##Making changes

After making any changes, build the command line tool with:
```
$ go build main/autotorrent_cli.go
```
