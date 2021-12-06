build.gitar:
	@echo "build in ${PWD}";go build gitar.go

build.image-gitar:
	docker build . -t gitar

install.gitar:
	@go build gitar.go && mv gitar ~/bin