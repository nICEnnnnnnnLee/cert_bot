ifeq ($(LANG),)
COMPILE_TIME = $(shell echo %date:~0,4%-%date:~5,2%-%date:~8,2%_%time:~0,2%:%time:~3,2%:%time:~6,2%)
else
COMPILE_TIME = $(shell date +"%Y-%m-%d_%H:%M:%S")
endif

ifeq ($(BOT_VERSION),)
BOT_VERSION=Unknown
endif

NAME=CertBot
BINDIR=bin
GOBUILD=CGO_ENABLED=0 go build -ldflags '-X main.version=$(BOT_VERSION) -X main.buildTime=$(COMPILE_TIME) -w -s -buildid='
# GOBUILD=CGO_ENABLED=0 go build -ldflags '-w -s -buildid='
# The -w and -s flags reduce binary sizes by excluding unnecessary symbols and debug info
# The -buildid= flag makes builds reproducible

#SET CGO_ENABLED=0
#SET GOOS=linux
#SET GOARCH=amd64 
#go build -ldflags ' -w -s -buildid='

all: linux-amd64 linux-386 linux-arm linux-arm64 macos-amd64 macos-arm64 win64 win32 freebsd-amd64

linux-arm:
	GOARCH=arm GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@
    
linux-arm64:
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@
    
linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-s390x:
	GOARCH=s390x GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@
    
linux-386:
	GOARCH=386 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@
    
freebsd-amd64:
	GOARCH=amd64 GOOS=freebsd $(GOBUILD) -o $(BINDIR)/$(NAME)-$@
    
macos-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

macos-arm64:
	GOARCH=arm64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

win64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe

win32:
	GOARCH=386 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe


test: test-linux-amd64 test-macos-amd64 test-macos-arm64 test-win64 test-win32

test-linux-amd64:
	GOARCH=amd64 GOOS=linux go test

test-macos-amd64:
	GOARCH=amd64 GOOS=darwin go test

test-macos-arm64:
	GOARCH=arm64 GOOS=darwin go test

test-win64:
	GOARCH=amd64 GOOS=windows go test

test-win32:
	GOARCH=386 GOOS=windows go test

build: linux-amd64 linux-386 linux-arm linux-arm64 win64 freebsd-amd64
	chmod +x $(BINDIR)/$(NAME)-*
	zip -m files.zip config.json static/* files/*
	cp files.zip $(BINDIR)/$(NAME)-linux-amd64.zip      && cd $(BINDIR) && zip -m $(NAME)-linux-amd64.zip       $(NAME)-linux-amd64 && cd ..
	cp files.zip $(BINDIR)/$(NAME)-linux-386.zip        && cd $(BINDIR) && zip -m $(NAME)-linux-386.zip         $(NAME)-linux-386 && cd ..
	cp files.zip $(BINDIR)/$(NAME)-linux-arm.zip        && cd $(BINDIR) && zip -m $(NAME)-linux-arm.zip         $(NAME)-linux-arm && cd ..
	cp files.zip $(BINDIR)/$(NAME)-linux-arm64.zip      && cd $(BINDIR) && zip -m $(NAME)-linux-arm64.zip       $(NAME)-linux-arm64 && cd ..
	cp files.zip $(BINDIR)/$(NAME)-freebsd-amd64.zip    && cd $(BINDIR) && zip -m $(NAME)-freebsd-amd64.zip     $(NAME)-freebsd-amd64 && cd ..
	cp files.zip $(BINDIR)/$(NAME)-win64.zip            && cd $(BINDIR) && zip -m $(NAME)-win64.zip             $(NAME)-win64.exe && cd ..

clean:
	rm $(BINDIR)/*

# Remove trailing {} from the release upload url
GITHUB_UPLOAD_URL=$(shell echo $${GITHUB_RELEASE_UPLOAD_URL%\{*})

upload:
	curl -H "Authorization: token $(GITHUB_TOKEN)" -H "Content-Type: application/zip" --data-binary @$(BINDIR)/$(NAME)-linux-386.zip        "$(GITHUB_UPLOAD_URL)?name=$(NAME)-linux-386.zip"
	curl -H "Authorization: token $(GITHUB_TOKEN)" -H "Content-Type: application/zip" --data-binary @$(BINDIR)/$(NAME)-linux-amd64.zip      "$(GITHUB_UPLOAD_URL)?name=$(NAME)-linux-amd64.zip"
	curl -H "Authorization: token $(GITHUB_TOKEN)" -H "Content-Type: application/zip" --data-binary @$(BINDIR)/$(NAME)-linux-arm.zip        "$(GITHUB_UPLOAD_URL)?name=$(NAME)-linux-arm.zip"
	curl -H "Authorization: token $(GITHUB_TOKEN)" -H "Content-Type: application/zip" --data-binary @$(BINDIR)/$(NAME)-linux-arm64.zip      "$(GITHUB_UPLOAD_URL)?name=$(NAME)-linux-arm64.zip"
	curl -H "Authorization: token $(GITHUB_TOKEN)" -H "Content-Type: application/zip" --data-binary @$(BINDIR)/$(NAME)-win64.zip            "$(GITHUB_UPLOAD_URL)?name=$(NAME)-win64.zip"
	curl -H "Authorization: token $(GITHUB_TOKEN)" -H "Content-Type: application/zip" --data-binary @$(BINDIR)/$(NAME)-freebsd-amd64.zip    "$(GITHUB_UPLOAD_URL)?name=$(NAME)-freebsd-amd64.zip"
