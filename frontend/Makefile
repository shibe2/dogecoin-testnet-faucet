deps:
	docker pull danielgtaylor/apisprout
	npm install -g browserify
	npm install -g bs58check
	
run:
	docker run --name MockAPI -d -p 8000:8000 danielgtaylor/apisprout https://gitlab.com/shibe1/faucet-backend/-/raw/master/openapi.yaml

build:
	echo "BUILD STARTING"

	echo "Cloning src"
	cp -r src dist

	echo "Adding commit hash"
	python3 commitHashInjector/main.py

	echo "Building JS file"
	browserify dist/mainUnbundled.js -o dist/main.js