gen-data:
	mkdir -p .tmp
	export GOTMPDIR=$(PWD)/.tmp && go run cmd/gen-data/main.go -out-dir=data

analyze:
	mkdir -p .tmp
	export GOTMPDIR=$(PWD)/.tmp && go run cmd/analyzer/main.go -data-dir=data

check:
	mkdir -p .tmp
	export GOTMPDIR=$(PWD)/.tmp && go run cmd/checker/main.go -data-dir=data

clean:
	rm -rf data
	rm -rf .tmp
