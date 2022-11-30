DOCKER_REPO := "wwonigkeit"
# RELEASE_TAG = $(shell v='$${RELEASE:-"latest"}$${RELEASE}'; echo "$${v%.*}")
RELEASE_TAG := 1.0

.PHONY: docker
docker: 
	@DOCKER_BUILDKIT=1 docker build -t ${DOCKER_REPO}/linode-receiver:${RELEASE_TAG} .
	@docker push ${DOCKER_REPO}/linode-receiver:${RELEASE_TAG}
	@echo "Make $@${RELEASE_TAG}: SUCCESS"

.PHONY: install
install: 
	kubectl apply -f kubernetes/install.yaml
	
.PHONY: uninstall
uninstall: 
	kubectl delete -f kubernetes/install.yaml