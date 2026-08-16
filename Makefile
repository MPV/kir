build:
	go build -o bin/ ./...

# go-approval-tests dropped its console reporter in v1.13.0, so a failing golden
# now reports only "received does not match approved" — no diff, no file names.
# The Systemout reporter prints both files and their contents. CI needs no such
# setting: GitHub Actions sets CI=true, which selects an equivalent reporter.
test:
	APPROVAL_TESTS_USE_REPORTER=SystemoutReporter go test ./... -v

# Accept the current output as the goldens. A failing approval test leaves what
# it produced beside its golden as *.received.txt; a passing one deletes its own,
# so the files left over are exactly the failures. Read what `make test` printed
# before running this, and run it after a full `make test` — a filtered
# `go test -run ...` leaves other tests' received files behind, and this accepts
# every one it finds.
approve:
	@for f in approvals/*.received.txt; do \
		[ -e "$$f" ] || continue; \
		mv "$$f" "$${f%.received.txt}.approved.txt"; \
		echo "approved $$(basename $${f%.received.txt})"; \
	done
