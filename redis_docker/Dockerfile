FROM redis:latest

# Copy custom redis.conf
COPY redis.conf /usr/local/etc/redis/redis.conf

# Set the default command to run Redis with the custom config
CMD ["redis-server", "/usr/local/etc/redis/redis.conf"]
