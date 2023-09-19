build:
	go build -o bin/quasar

run: build
	./bin/quasar
