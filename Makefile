build.gitar:
	@echo "build in ${PWD}";go build gitar.go

build.image-gitar:
	ko publish -B gitar.go

install.gitar:
	@go build gitar.go && mv gitar ~/bin