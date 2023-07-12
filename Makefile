BUILDVERSION:=latest
DOCKERIMAGE:=stackmap-consumer:$(BUILDVERSION)

.PHONY: kind-load
.PHONY: build-docker

#build:
#	go build -o bin/restapi cmd/api/main.go

#run:
#	go run cmd/api/main.go

#test:
#	go test -v ./test/...

build-docker: build
	docker build . -t $(DOCKERIMAGE)

#run-docker: build-docker
	#export POSTGRES_PASSWORD=$(kubectl get secret --namespace default postgresql -o jsonpath="{.data.postgres-password}" | base64 -d) && docker run --rm -it --network host -e "PGPASSWORD=$POSTGRES_PASSWORD" -v $(pwd)/src:/src -v $(pwd)/data:/app/data test /bin/sh

port-forward:
	kubectl port-forward svc/postgresql 5432:5432

kind-delete: 
	kubectl delete deployments/stackmap-consumer

kind-load: build-docker
	kind load docker-image $(DOCKERIMAGE)

kind-deploy: kind-load
	kubectl apply -f deployment.yaml 

#swagger-build:
#	swagger generate spec -i ./swagger/swagger_base.yaml -o ./swagger.yaml

swagger-serve:
	cd swagger && swagger serve --flatten --port=9009 -F=swagger swagger.yaml
