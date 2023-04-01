env GOOS=windows GOARCH=386 go build -o ./build/sebulk_win32.exe .
env GOOS=windows GOARCH=amd64 go build -o ./build/sebulk_win64.exe .
env GOOS=linux GOARCH=amd64 go build -o ./build/sebulk_linux64 .
go build -o ./build/sebulk_macos .