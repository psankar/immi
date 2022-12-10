funnel:
	cd backend && go build -o cmd/funnel/funnel cmd/funnel/funnel.go

test:
	cd backend && ginkgo -r -v -race --trace --coverprofile=.coverage-report.out ./...
