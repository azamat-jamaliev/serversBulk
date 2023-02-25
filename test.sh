cd ./test
docker build -t sebulk_test .
docker run -d -p 22:22 sebulk_test
cd ..
go test