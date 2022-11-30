DOCKER_REPO := "localhost:5000"
RELEASE_TAG = $(shell v='$${RELEASE:-"latest"}$${RELEASE}'; echo "$${v%.*}")

.PHONY: docker
docker: 
	@DOCKER_BUILDKIT=1 docker build -t ${DOCKER_REPO}/equinix-listener:${RELEASE_TAG} .
	@docker push ${DOCKER_REPO}/equinix-listener:${RELEASE_TAG}
	@echo "Make $@${RELEASE_TAG}: SUCCESS"

.PHONY: install
install: 
	kubectl apply -f kubernetes/install.yaml
	
.PHONY: uninstall
uninstall: 
	kubectl delete -f kubernetes/install.yaml