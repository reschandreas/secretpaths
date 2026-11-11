.PHONY: dev dev-server dev-visualizer test lint

dev-server:
	cd server && go run .

dev-visualizer:
	cd visualizer && npm run dev

test:
	cd server && go test ./...
	cd visualizer && npm run check

lint:
	cd visualizer && npm run lint

dev:
	@echo "Run 'make dev-server' and 'make dev-visualizer' in separate terminals."