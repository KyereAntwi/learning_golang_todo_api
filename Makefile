.PHONY: run build clean

run:
	cd src/api && go run .

build:
	cd src/api && go build -o ../../bin/api .

clean:
	rm -rf bin/
