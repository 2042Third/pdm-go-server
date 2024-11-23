run

`docker build -t custom-redis .`

`docker run -p 6379:6379 -v redis-data:/data custom-redis`
