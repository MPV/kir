build:
	go build -o bin/ ./...

# go-approval-tests dropped its console reporter in v1.13.0, so a failing golden
# now reports only "received does not match approved" — no diff, no file names.
# The Systemout reporter prints both files and their contents. CI needs no such
# setting: GitHub Actions sets CI=true, which selects an equivalent reporter.
test:
	APPROVAL_TESTS_USE_REPORTER=SystemoutReporter go test ./... -v
