# SeBulk
The sebulk allows to 
- quick serach in the logs of different nodes and servers when GrayLog is not available
- download all logs (in case if you need to attach logs from all server to ticket for DEV team)
- upload file (if you have hotfix which should be alloaded to all nodes / servers)
- and execute commands on the servers simalteniusly when Ansible is not available

UI has been added (not required to learn cli parameters any more):
![enter image description here](./doc/pictures/mainPage.png)

## Execute via CLI
### to execute command
```
 .\sebulk.exe -c .\win_sebulk_config_PPT.json --servers SERVER_GROUP_NAME -e "ls -l /var/tmp"
```
### for search
```
./sebulk -s "\[ERROR" > ~/Downloads/output.txt
```
### for logs download
```
./sebulk --servers SERVER_GROUP_NAME -c ./config/sebulk_config_SVT.json  -d ~/Downloads
```

### upload file to servers
```
./sebulk -c ./config/sebulk_config_SVT.json  -u ~/Downloads/file_to_upload.txt
```
__NOTE:__ the file will be uploaded to /var/tmp folder (if folder does not exist uploading will fail for the server)


## Config file examples
### If connection should be performed via Bastion server
#### Bastion server password authentication is used
```yaml
{
    "servers": [
        {
            "name": "SERVER_GROUP_NAME",
            "description": "",
            "logFolders": ["/var/tmp/logs"],
            "logFilePattern": "*.log",
            "BastionServer": "192.XXX.XXX.1",
            "BastionLogin": "YourLoginforBastion",
            "BastionPassword": "YourPasswordforBastion",
            "login": "YourLoginToTheServerGroup",
            "passowrd": "YourLoginToTheServerGroup",
            "ipAddresses": [
                "172.XXX.XXX.XX1",
                "172.XXX.XXX.XX2"
            ]
        }
    ]
}
```

#### Bastion server Public/Private Key File authentication is used
```yaml
{
    "servers": [
        {
            "name": "SERVER_GROUP_NAME",
            "description": "",
            "logFolders": ["/var/tmp/logs"],
            "logFilePattern": "*.log",
            "BastionServer": "192.XXX.XXX.1",
            "BastionLogin": "YourLoginforBastion",
            "BastionIdentityFile": "/Users/username/.ssh/key_rsa",
            "login": "YourLoginToTheServerGroup",
            "passowrd": "YourLoginToTheServerGroup",
            "ipAddresses": [
                "172.XXX.XXX.XX1",
                "172.XXX.XXX.XX2"
            ]
        }
    ]
}
```
### Config without Bastion server
```yaml
{
    "servers": [
        {
            "name": "SERVER_GROUP_NAME",
            "description": "",
            "logFolders": ["/var/tmp/logs"],
            "logFilePattern": "*.log",
            "login": "YourLoginToTheServerGroup",
            "passowrd": "YourLoginToTheServerGroup",
            "ipAddresses": [
                "172.XXX.XXX.XX1",
                "172.XXX.XXX.XX2"
            ]
        }
    ]
}
```

## Build

if your you're building in the same platform you can use simple:
```sh
go build -o ./build .
```
for cross platform building you can use:
```sh
env GOOS=target-OS GOARCH=target-architecture go build 
```
more details in: https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04


### Cross platform: Build for Linux
```sh
env GOOS=linux GOARCH=amd64 go build -o ./build .
```
### Cross platform: Build for Windows
```sh
env GOOS=windows GOARCH=386 go build -o ./build/sebulk.exe .
env GOOS=windows GOARCH=amd64 go build -o ./build/sebulk.exe .
```