export containerId=$(docker ps -a -q --filter ancestor=sebulk_test --format="{{.ID}}")
if [ -z "$containerId" ] 
then
    docker build -t sebulk_test ./test/
    docker run -d -p 22:22 sebulk_test
    export containerId=$(docker ps -a -q --filter ancestor=sebulk_test --format="{{.ID}}")
fi
go test ./...
docker stop $containerId
docker rm $containerId