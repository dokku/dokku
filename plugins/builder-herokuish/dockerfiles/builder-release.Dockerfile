ARG APP_IMAGE
FROM $APP_IMAGE

COPY 00-global-env.sh 01-app-env.sh /app/.profile.d/
