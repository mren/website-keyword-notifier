# Alert

> Notifies via Email if a certain word is found on a website.

## Build

```bash
go build
```

## Configuration

Environment variables are used to configure the cronjob.

```bash
CONFIG_URL=file://example-configuration.json \
MAILGUN_DOMAIN=example.com \
MAILGUN_KEY=secret=key \
./alert
```

## Deploy

This project can be deployed on heroku and executed via the Heroku Scheduler on a free dyno.

```
heroku apps:create
heroku config:set CONFIG_URL=https://...
heroku addons:create scheduler:standard
heroku addons:open scheduler # add `bin/alert` to scheduler
```
