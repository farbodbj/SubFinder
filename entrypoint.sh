docker build . -t subfinder
docker run -v $(pwd)/data:/subfinder/data subfinder