# Install xxhsum from https://github.com/Cyan4973/xxHash
IMMI_MIGRATIONS_TAG := $(shell tar -cf - --sort=name db-migrations/sql | xxhsum | cut -d ' ' -f1)

godeps:
	go install github.com/onsi/ginkgo/v2/ginkgo@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.2

funnel:
	cd backend && go build -o cmd/funnel/funnel cmd/funnel/funnel.go

test:
	cd backend && ginkgo -r -v -race --trace --coverprofile=.coverage-report.out ./...

env:
	# docker pull postgres:14.6
	# Uncomment above line if you do not have the image locally
	docker tag postgres:14.6 localhost:5001/postgres:14.6
	# --------------------

	./kind-with-registry.sh
	kubectl cluster-info --context kind-kind
	kubectl create ns immi
	kind load docker-image localhost:5001/postgres:14.6

	kubectl apply -f ./immi-dev-env.yaml
	# Wait for a few seconds to let the pods come up
	sleep 5
	# Wait until the database is up and running
	kubectl wait --for=condition=ready -n immi pod -l app=postgres --timeout=60s
	nohup kubectl port-forward -n immi deployment/postgres 5432 &
	kubectl get pods -n immi

env-clean:
	kind delete cluster

migrate:
	cd db-migrations && \
		docker build -t localhost:5001/immi-migrations:$(IMMI_MIGRATIONS_TAG) -f Dockerfile .
	kind load docker-image localhost:5001/immi-migrations:$(IMMI_MIGRATIONS_TAG)
	cd db-migrations && \
		# Bash scripter's Helm \
		sed 's/IMMI_MIGRATIONS_TAG/$(IMMI_MIGRATIONS_TAG)/' db-migrations-dev.yaml | \
		kubectl apply -n immi -f -
	# Wait for migrations to get completed
	kubectl wait --for=condition=complete -n immi job db-migrations --timeout=60s
	kubectl get pods -n immi

devdeploy:
	cd backend && KO_DOCKER_REPO=kind.local ko apply -f backend-dev.yaml
	kubectl get pods -n immi
