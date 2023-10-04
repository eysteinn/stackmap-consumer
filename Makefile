BUILDVERSION:=latest
DOCKERIMAGE:=stackmap-consumer:$(BUILDVERSION)

.PHONY: kind-load
.PHONY: build-docker


build-docker: build
	docker build . -t $(DOCKERIMAGE)

build-docker-base: 
	docker build . -t stackmap-consumer:base --target build_base

port-forward:
	kubectl port-forward svc/postgresql 5432:5432

kind-delete: 
	kubectl delete deployments/stackmap-consumer

kind-load: build-docker
	kind load docker-image $(DOCKERIMAGE)

kind-deploy: kind-load
	kubectl apply -f deployment.yaml 

k3s-deploy:
	docker save $(DOCKERIMAGE) | sudo k3s ctr images import -

k-delete:
	kubectl delete deploy stackmap-consumer

k-deploy:
	kubectl apply -f https://raw.githubusercontent.com/eysteinn/stackmap-consumer/main/deployment.yaml


swagger-serve:
	cd swagger && swagger serve --flatten --port=9009 -F=swagger swagger.yaml
