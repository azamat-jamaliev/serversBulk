env GOOS=windows GOARCH=386 go build -o ./build/win32/sebulk.exe .
env GOOS=windows GOARCH=amd64 go build -o ./build/win64/sebulk.exe .
env GOOS=linux GOARCH=amd64 go build -o ./build/linux64/ .
go build -o ./build/macos/ .