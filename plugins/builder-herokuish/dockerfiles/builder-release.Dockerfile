ARG APP_IMAGE
FROM $APP_IMAGE

ARG DOKKU_APP_USER herokuishuser
COPY --chown=$DOKKU_APP_USER 00-global-env.sh 01-app-env.sh /app/.profile.d/
RUN find "/app" \( \! -user "$DOKKU_APP_USER" -o \! -group "$DOKKU_APP_USER" \) -print0 | xargs -0 -r chown "$DOKKU_APP_USER:$DOKKU_APP_USER"
USER $DOKKU_APP_USER
ENV HEROKUISH_SETUIDGUID false
