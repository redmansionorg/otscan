.PHONY: build run dev clean frontend

build: frontend
	go build -o otscan ./cmd/otscan

run: build
	./otscan --config config.yaml

dev:
	go run ./cmd/otscan --config config.yaml

frontend:
	cd web && npm install && npm run build

clean:
	rm -f otscan
	rm -rf web/dist web/node_modules
