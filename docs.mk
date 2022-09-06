.PHONY: docs-build-image
docs-build-image:
	 docker build -f docs/_build/Dockerfile --progress plain -t app/mkdocs .

.PHONY: docs-build
docs-build:
	@docker run --rm -p 3487:3487 -v "$(PWD):/usr/src/source" --name docs-serve app/mkdocs build

.PHONY: docs-serve
docs-serve:
	docker run --rm -p 3487:3487 -v "$(PWD):/usr/src/source" --name docs-serve app/mkdocs serve

.PHONY: docs-deps
docs-deps:
	@docker run --rm -p 3487:3487 -v "$(PWD):/usr/src/app" app/mkdocs deps

.PHONY: docs-enter
docs-enter:
	docker exec -it docs-serve bash
