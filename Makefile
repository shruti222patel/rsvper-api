build:
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -o bin/bot bot/main.go

.PHONY: clean
clean:
	rm -rf ./bin ./vendor Gopkg.lock

.PHONY: deploy
deploy: clean build
	sls deploy --verbose

.PHONY: test
test:
	export GOOGLE_API_CREDS=<insert google api creds>
	go test ./...