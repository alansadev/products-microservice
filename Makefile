COVER_PKGS = "./handlers,./middleware,./models"

COVER_PROFILE = coverage

default: help

test-coverage:
	@echo "--> Generating test coverage report..."
	go test -coverpkg=$(COVER_PKGS) -coverprofile=$(COVER_PROFILE) ./...
	@echo "--> Open coverage in browser..."
	go tool cover -html=$(COVER_PROFILE)

help:
	@echo ""
	@echo "Available commands:"
	@echo " make test-coverage    Generates and displays the test coverage report."
	@echo ""