BUILD_DIR=build/
BUILD_OS=linux darwin freebsd
BUILD_ARCH=386 amd64
BUILD_ARM=5 6 7

clean:
	rm cmd/agent/build/* || true

build_all: clean
	mkdir -p cmd/agent/build
	cd cmd/agent && for OS in $(BUILD_OS) ; do \
		for ARCH in $(BUILD_ARCH) ; do \
			echo Compiling $$OS $$ARCH ; \
			GOARCH=$$ARCH GOOS=$$OS go build -o "build/agent-$$OS-$$ARCH" ; \
		done ; \
		for ARM in $(BUILD_ARM) ; do \
			echo Compiling $$OS armv$$ARM ; \
			GOARCH=arm GOOS=$$OS go build -o "build/agent-$$OS-armv$$ARM" ; \
		done \
	done

agent:
	cd cmd/agent && go install
