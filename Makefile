Workspace := $(shell pwd)
GoInstallCmd := env GOPATH=$(Workspace) go install -v log_analysis/cmd
Commands = feedback-count

# default target
bin: $(Commands)

.PHONY: bin clean

$(Commands):
	$(GoInstallCmd)/$@

go_vendor:
	cd src/log_analysis; env GOPATH=$(Workspace) dep ensure -vendor-only

clean:
	-rm bin/*
