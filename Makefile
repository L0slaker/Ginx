.PHONY: mock
mock:
		@mockgen -source=internal\ratelimit\types.go -destination=internal\ratelimit\mocks\ratelimit.mock.go -package=limitmocks
		@go mod tidy
