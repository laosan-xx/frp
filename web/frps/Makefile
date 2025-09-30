.PHONY: dist build preview lint

build:
	@npm run build
	@echo "Copying build files to assets/frps/static..."
	@rm -rf ../../assets/frps/static/*
	@cp -r dist/* ../../assets/frps/static/
	@echo "Build files copied successfully!"

dev:
	@npm run dev

preview:
	@npm run preview

lint:
	@npm run lint
