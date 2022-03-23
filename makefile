BINARY=tenant-terraform-generator

build:
	go build -o ${BINARY}

run:
	go run main.go