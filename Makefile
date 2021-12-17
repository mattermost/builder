SHELL:=/usr/bin/env bash
.PHONY:  build release publish
build:
	mkdir tmp || :
	rm -f tmp/builder_linux_amd64 tmp/builder_darwin_amd64 
	env CGO_ENABLED=0 go build -o tmp/mmbuilder_linux_amd64 cmd/mmbuild/main.go
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o tmp/mmbuilder_darwin_amd64 cmd/mmbuild/main.go

publish:
	if [ -z "${TAG}" ]; then echo "Tag is not set"; exit 1; fi
	#go get k8s.io/release/cmd/publish-release
	#go run cmd/publish-release/main.go github \
	publish-release github --nomock \
		--repo=mattermost/builder --tag=${TAG} \
		--asset="tmp/mmbuilder_linux_amd64:Mattermost Builder for linux/amd64"  \
		--asset="tmp/mmbuilder_darwin_amd64:Mattermost Builder for darwin/amd64" 

release: build publish

##@ Helpers
