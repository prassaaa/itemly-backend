.PHONY: test test-verbose test-cover test-cover-html k6-auth k6-ratelimit k6

test:
	go test ./... -count=1

test-verbose:
	go test ./... -v -count=1

test-cover:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out

test-cover-html:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out

k6-auth:
	k6 run k6/auth_load.js

k6-ratelimit:
	k6 run k6/ratelimit.js

k6: k6-auth k6-ratelimit
