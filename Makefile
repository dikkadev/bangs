# Change these variables as necessary.
main_package_path = .
binary_name = bangs

VERSION = $(shell git describe --tags --dirty)

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: no-dirty
no-dirty:
	@test -z "$(shell git status --porcelain)"


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: run quality control checks
.PHONY: audit
audit: test
	go mod tidy -diff
	go mod verify
	test -z "$(shell gofmt -l .)" 
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## tidy: tidy modfiles and format .go files
.PHONY: tidy
tidy:
	go mod tidy -v
	go fmt ./...

## build: build the application
.PHONY: build
build:
	# Include additional build steps, like TypeScript, SCSS or Tailwind compilation here...
	go build -ldflags "-X main.version=`git describe --tags --dirty`" -o=/tmp/bin/${binary_name} ${main_package_path}

## run: run the  application
.PHONY: run
run: build
	/tmp/bin/${binary_name} -vwab bangs.yaml

## run/live: run the application with reloading on file changes
.PHONY: run/live
run/live:
	go run github.com/cosmtrek/air@v1.43.0 \
		--build.cmd "make build" --build.bin "/tmp/bin/${binary_name} -v -w -a -b bangs.yaml" --build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, jpeg, jpg, gif, png, bmp, svg, webp, ico" \
		--misc.clean_on_exit "true"


# ==================================================================================== #
# OPERATIONS
# ==================================================================================== #

## push: push changes to the remote Git repository
.PHONY: push
push: confirm audit no-dirty
	git push

# ==================================================================================== #
# TEMPL
# ==================================================================================== #

## generate go files from templ files
.PHONY: templ
templ:
	find . -type f \( -name "*.templ" -or -name "*.css" -or -name "*.js" \) | entr -r  templ generate

## add templ to project
.PHONY: templinit
templinit:
	go get github.com/a-h/templ
	mkdir -p ./web/
	touch ./web/index.templ


# ==================================================================================== #
# DOCKER
# ==================================================================================== #

## docker/build: build the Docker image
.PHONY: docker/build
docker/build:
	docker build --build-arg VERSION=$(VERSION) -t bangs .

## docker/run: run the Docker container
.PHONY: docker/run
docker/run:
	docker run -it --rm -p 8080:8080 -e BANGS_BANGFILE=bangs.yaml -e BANGS_VERBOSE=true -e BANGS_WATCH=true -v $(PWD)/bangs.yaml:/app/bangs.yaml bangs
