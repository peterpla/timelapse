all: bin/timelapse
.PHONY: bin/timelapse

bin/timelapse:
	@docker build . --target bin \
	--output bin/