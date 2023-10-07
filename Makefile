.PHONY: mock
mock:
		@mockgen -source=internal\ratelimit\types.go -destination=internal\ratelimit\mocks\ratelimit.mock.go -package=limitmocks
		@mockgen -destination=middlewares\redislimit\redismocks\cmdable.mock.go -package=redismocks github.com/redis/go-redis/v9 Cmdable
		@go mod tidy

#  go install go.uber.org/mock/mockgen@latest