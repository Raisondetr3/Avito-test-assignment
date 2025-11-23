.PHONY: build up down integration-test lint lint-fix clean

build:
	go build -o bin/service ./cmd/service

up:
	docker-compose up --build

down:
	docker-compose down

integration-test:
	@docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true
	@docker-compose -f docker-compose.test.yml up -d
	@sleep 3
	@TEST_DB_DSN="postgres://reviewer_user:reviewer_pass@localhost:5433/pr_reviewer_test?sslmode=disable" \
		go test -v ./integration_test/... -timeout 30s; \
		TEST_RESULT=$$?; \
		docker-compose -f docker-compose.test.yml down -v; \
		exit $$TEST_RESULT

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

clean:
	rm -rf bin/
	docker-compose down -v
	docker-compose -f docker-compose.test.yml down -v
