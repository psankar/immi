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

devdeploy:
	cd backend && KO_DOCKER_REPO=kind.local ko apply -f backend-dev.yaml
	kubectl get pods -n immi
