Workspace := $(shell pwd)
GoInstallCmd := env GOPATH=$(Workspace) go install -v analysis-tools
Commands = feedback-count

.PHONY: bin go-vender

bin: $(Commands)

$(Commands):
	$(GoInstallCmd)/$@

go-vendor:
	cd src/analysis-tools; env GOPATH=$(Workspace) dep ensure -vendor-only
