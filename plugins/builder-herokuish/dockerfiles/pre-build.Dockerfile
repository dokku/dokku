ARG APP_IMAGE
FROM $APP_IMAGE

COPY .env.d /tmp/env
COPY .env /app/.env
