FROM iron/go
COPY sentinel /usr/bin/
CMD ["sentinel"]
