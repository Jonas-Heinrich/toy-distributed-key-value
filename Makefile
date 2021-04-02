build-kv:
	go build -o build/toy-distributed-kv

run-kv: build-kv
	./build/toy-distributed-kv

clean:
	find build ! -name '.gitignore' -type f -exec rm -f {} +
