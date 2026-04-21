.PHONY: run build clean

run:
	cd src/workers && go run . & cd src/api && go run . & wait

build:
	cd src/api && go build -o ../../bin/api .

clean:
	rm -rf bin/
