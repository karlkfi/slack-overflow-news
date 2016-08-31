# slackstack
A bot for reporting new Stack Overflow questions into Slack

# Usage

```
cp .env.example .env
```

Fill out missing config vars in `.env` (like token).

```
docker run -it --env-file "$(pwd)/.env" karlkfi/slackstack
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
