.PHONY: generate

build:
	@go build -o ./bin/webshell main.go
	@echo "[OK] App binary was created!"

build-all:
	GOOS=linux go build -o bin/webshell-linux
	GOOS=darwin go build -o bin/webshell-mac
	GOOS=windows go build -o bin/webshell.exe
	@echo "[OK] App binary was created for all platforms!"

run:
ifneq ("$(wildcard $(./bin/webshell))","")
	@./bin/webshell
else
	@go build -o ./bin/webshell main.go
	@./bin/webshell
endif
