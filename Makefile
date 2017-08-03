BUILD_DIR=build/
BUILD_OS=linux darwin freebsd
BUILD_ARCH=386 amd64

build_all:
	cd cmd/agent && for OS in $(BUILD_OS) ; do \
		for ARCH in $(BUILD_ARCH) ; do \
			echo $$OS $$ARCH ; \
			GOARCH=$$ARCH GOOS=$$OS go build -o "agent-$$OS-$$ARCH" ; \
		done \
	done

agent:
	cd cmd/agent && go install
