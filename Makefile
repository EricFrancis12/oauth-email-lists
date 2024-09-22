test:
	go test -v ./...

build_serverless:
	npm install
	env GOOS=linux go build -ldflags="-s -w" -o bootstrap .

deploy_serverless: build_serverless
	serverless deploy --stage prod
