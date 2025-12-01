.PHONY: test test-chaosnet test-local test-verbose deps clean help \
        docker-build docker-test docker-smoke docker-account docker-clean \
        docker-format ci-test ci-format

# Default test server
INTERX_URL ?= http://3.123.154.245:11000
TEST_ADDRESS ?= kira143q8vxpvuykt9pq50e6hng9s38vmy844n8k9wx

help:
	@echo "Interx Integration Tests"
	@echo ""
	@echo "Docker Commands (Recommended):"
	@echo "  make docker-build   Build the test container"
	@echo "  make docker-test    Run all tests in container"
	@echo "  make docker-smoke   Run smoke tests in container"
	@echo "  make docker-account Run account tests in container"
	@echo "  make docker-format  Run format validation tests"
	@echo "  make docker-clean   Remove test containers and images"
	@echo ""
	@echo "CI Commands (used by GitHub Actions):"
	@echo "  make ci-test        Run full test suite with summary"
	@echo "  make ci-format      Run format validation with summary"
	@echo ""
	@echo "Local Commands:"
	@echo "  make deps           Install dependencies"
	@echo "  make test           Run all integration tests"
	@echo "  make test-chaosnet  Run tests against chaosnet"
	@echo "  make test-local     Run tests against local instance"
	@echo "  make test-single    Run a single test (use TEST=TestName)"
	@echo "  make clean          Clean test cache"
	@echo ""
	@echo "Environment variables:"
	@echo "  INTERX_URL          Override the test server URL"
	@echo "  TEST_ADDRESS        Override the test account address"
	@echo "  VALIDATOR_ADDRESS   Override the validator address"
	@echo "  DELEGATOR_ADDRESS   Override the delegator address"

# =============================================================================
# Docker Commands
# =============================================================================

docker-build:
	docker build -t interx-integration-tests .

docker-test: docker-build
	docker run --rm --network host \
		-e INTERX_URL=$(INTERX_URL) \
		-e TEST_ADDRESS=$(TEST_ADDRESS) \
		interx-integration-tests

docker-smoke: docker-build
	docker run --rm --network host \
		-e INTERX_URL=$(INTERX_URL) \
		-e TEST_ADDRESS=$(TEST_ADDRESS) \
		interx-integration-tests \
		go test -v -count=1 -run "TestInterxStatus|TestKiraStatus" ./...

docker-account: docker-build
	docker run --rm --network host \
		-e INTERX_URL=$(INTERX_URL) \
		-e TEST_ADDRESS=$(TEST_ADDRESS) \
		interx-integration-tests \
		go test -v -count=1 -run "TestQuery(Accounts|Balances)" ./...

docker-single: docker-build
	docker run --rm --network host \
		-e INTERX_URL=$(INTERX_URL) \
		-e TEST_ADDRESS=$(TEST_ADDRESS) \
		interx-integration-tests \
		go test -v -count=1 -run "$(TEST)" ./...

docker-compose-test:
	docker-compose up --build --abort-on-container-exit integration-tests

docker-compose-smoke:
	docker-compose up --build --abort-on-container-exit smoke-tests

docker-clean:
	docker rmi -f interx-integration-tests 2>/dev/null || true
	docker-compose down --rmi local 2>/dev/null || true

docker-format: docker-build
	docker run --rm --network host \
		-e INTERX_URL=$(INTERX_URL) \
		-e TEST_ADDRESS=$(TEST_ADDRESS) \
		interx-integration-tests \
		go test -v -count=1 -run "ResponseFormat" ./...

# =============================================================================
# CI Commands (used by GitHub Actions)
# =============================================================================

ci-test: docker-build
	@echo "Running full integration test suite..."
	docker run --rm --network host \
		-e INTERX_URL=$(INTERX_URL) \
		-e TEST_ADDRESS=$(TEST_ADDRESS) \
		interx-integration-tests \
		go test -v -count=1 -timeout=10m ./... 2>&1 | tee test-output.log
	@echo ""
	@echo "=== Test Summary ==="
	@grep -c "^--- PASS" test-output.log | xargs -I{} echo "Passed: {}"
	@grep -c "^--- FAIL" test-output.log | xargs -I{} echo "Failed: {}"

ci-format: docker-build
	@echo "Running miro format validation tests..."
	docker run --rm --network host \
		-e INTERX_URL=$(INTERX_URL) \
		-e TEST_ADDRESS=$(TEST_ADDRESS) \
		interx-integration-tests \
		go test -v -count=1 -run "ResponseFormat" ./... 2>&1 | tee format-test-output.log
	@echo ""
	@echo "=== Format Validation Summary ==="
	@grep -c "^--- PASS" format-test-output.log | xargs -I{} echo "Passed: {}"
	@grep -c "^--- FAIL" format-test-output.log | xargs -I{} echo "Failed: {}"

# =============================================================================
# Local Commands
# =============================================================================

deps:
	go mod download
	go mod tidy

test: deps
	INTERX_URL=$(INTERX_URL) TEST_ADDRESS=$(TEST_ADDRESS) go test -v -count=1 ./...

test-chaosnet: deps
	INTERX_URL=http://3.123.154.245:11000 TEST_ADDRESS=kira143q8vxpvuykt9pq50e6hng9s38vmy844n8k9wx go test -v -count=1 ./...

test-local: deps
	TEST_ENV=local go test -v -count=1 ./...

test-verbose: deps
	INTERX_URL=$(INTERX_URL) TEST_ADDRESS=$(TEST_ADDRESS) go test -v -count=1 -race ./...

test-single: deps
	INTERX_URL=$(INTERX_URL) TEST_ADDRESS=$(TEST_ADDRESS) go test -v -count=1 -run $(TEST) ./...

# Run tests by category
test-account: deps
	INTERX_URL=$(INTERX_URL) go test -v -count=1 -run "TestQuery(Accounts|Balances)" ./...

test-transactions: deps
	INTERX_URL=$(INTERX_URL) go test -v -count=1 -run "Test(Transaction|QueryBlocks|QueryTransactions)" ./...

test-validators: deps
	INTERX_URL=$(INTERX_URL) go test -v -count=1 -run "Test(QueryValidators|Consensus|DumpConsensus)" ./...

test-status: deps
	INTERX_URL=$(INTERX_URL) go test -v -count=1 -run "Test(KiraStatus|InterxStatus|Dashboard|Metadata)" ./...

clean:
	go clean -testcache
