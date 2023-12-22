FROM alpine:latest
RUN apk -U add yt-dlp
RUN apk -U add ffmpeg
RUN apk add --no-cache libc6-compat 

RUN mkdir /app
WORKDIR /app

COPY /bin .
COPY .env .

CMD ["./main"]
