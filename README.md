# slack-overflow-news

A bot for reporting new Stack Overflow questions into Slack

# Usage

Create a `.env` file:

```
curl -o .env https://raw.githubusercontent.com/karlkfi/slack-overflow-news/master/.env.example
```

Fill out missing config vars in `.env` (like token).

Run slack-overflow-news in Docker:

```
docker run -it --env-file "$(pwd)/.env" karlkfi/slack-overflow-news
```

# Build

Build binaries:

```
ci/build.sh
```

Build Docker image:

```
ci/build-image.sh
```
