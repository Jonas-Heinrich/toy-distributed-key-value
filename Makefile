SUDO_PREFIX=sudo

# Careful: I haven't tested the setup steps below
setup:
	wget https://golang.org/doc/install?download=go1.16.3.linux-amd64.tar.gz
	rm -rf /usr/local/go && tar -C /usr/local -xzf go1.14.3.linux-amd64.tar.gz
	echo "export PATH=$PATH:/usr/local/go/bin" > ~/.bashrc
	$(SUDO_PREFIX) apt-get update && sudo apt-get install -y docker docker-compose

build-kv:
	go build -o build/toy-distributed-kv

build-docker:
	$(SUDO_PREFIX) docker build . -t toy-distributed-key-value

run-kv: build-kv
	./build/toy-distributed-kv

run-dc: build-docker
	$(SUDO_PREFIX) docker-compose up --scale follower=3 #--abort-on-container-exit

clean:
	find build ! -name '.gitignore' -type f -exec rm -f {} +
