ARG APP_IMAGE
FROM $APP_IMAGE

COPY --chown=$DOKKU_APP_USER .env.d /tmp/env
COPY --chown=$DOKKU_APP_USER .env /app/.env
