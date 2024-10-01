BIN_FILE_PATH := ./oauth-email-lists

build:
	go build -o $(BIN_FILE_PATH)

run: build
	$(BIN_FILE_PATH)

test:
	go test -v ./...

build_serverless:
	npm install
	env GOOS=linux go build -ldflags="-s -w" -o bootstrap .

deploy_serverless: build_serverless
	serverless deploy --stage prod

create-env:
	./scripts/create_env_file.sh
