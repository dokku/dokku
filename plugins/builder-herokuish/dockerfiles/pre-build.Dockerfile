ARG APP_IMAGE
FROM $APP_IMAGE

ARG DOKKU_APP_USER herokuishuser
COPY --chown=$DOKKU_APP_USER .env.d /tmp/env
COPY --chown=$DOKKU_APP_USER .env /app/.env
