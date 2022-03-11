before.build:
	go mod tidy && go mod download

build.gitar:
	@echo "build in ${PWD}";go build gitar.go

build.gitar.image:
	docker build . -t ariary/gitar

install.gitar:
	@go build gitar.go && mv gitar ~/bin
