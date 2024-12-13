.PHONY: build-image
build-image:
	@docker build \
		--no-cache \
		-t xpdb-docs ../.github/actions/mkdocs

.PHONY: build-docs
build-docs: build-image
	@docker run --rm -it \
		-p 8000:8000 \
		-v ${PWD}:/docs \
		--entrypoint /bin/bash   \
		xpdb-docs \
		-c "pip install -r /docs/requirements.txt && mkdocs build --config-file mkdocs.yml"

.PHONY: live-docs
live-docs: build-image ## Build and launch a local copy of the documentation website in http://localhost:8000
	@docker run --rm -it \
		-p 8000:8000 \
		-v ${PWD}:/docs \
		--entrypoint /bin/bash   \
		xpdb-docs \
		-c "pip install -r /docs/requirements.txt && mkdocs serve --dev-addr=0.0.0.0:8000"
