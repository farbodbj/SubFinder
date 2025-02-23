# make sure image is tagged as subfinder
docker build . -t subfinder
docker run -v $(pwd)/data:/subfinder/data subfinder
