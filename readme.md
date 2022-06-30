# serversBulk
The serversBulk allows to 
- quick serach in the logs of different nodes and servers when Grafana is not available :-) 
- download all logs
- upload filed 
- and execute commands on the servers simalteniusly when Ansible is not available.

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
env GOOS=windows GOARCH=386 go build -o ./build/serversBulk.exe .
```

## Execute
### to execute command
```
 .\serversBulk.exe -c .\win_serversBulk_config_PPT.json --servers SERVER_GROUP_NAME -e "ls -l /var/tmp"
```
### for search
```
./serversBulk -s "\[ERROR" > ~/Downloads/output.txt
```
### for logs download
```
./serversBulk --servers SERVER_GROUP_NAME -c ./config/serversBulk_config_SVT.json  -d ~/Downloads
```

```
./serversBulk --servers SERVER_GROUP_NAME -e "curl -v -g http://localhost:8080/health"
```

## Config file examples
### If connection should be performed via Bastion server
#### Bastion server password authentication is used
```yaml
{
    "servers": [
        {
            "name": "SERVER_GROUP_NAME",
            "description": "",
            "logFolder": "/var/tmp/logs",
            "logFilePattern": "*.log",
            "searchInSubfolders": true,
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
            "logFolder": "/var/tmp/logs",
            "logFilePattern": "*.log",
            "searchInSubfolders": true,
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

