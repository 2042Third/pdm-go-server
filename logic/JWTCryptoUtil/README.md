GOOS=darwin GOARCH=arm64 go build -o keyGen_macos_arm64 ./cryptoUtil 

GOOS=linux GOARCH=amd64 go build -o keyGen_linux_amd64 ./cryptoUtil

GOOS=linux GOARCH=arm64 go build -o keyGen_linux_arm64 ./cryptoUtil