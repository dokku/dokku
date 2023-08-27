ARG APP_IMAGE
FROM $APP_IMAGE

RUN printf '#!/usr/bin/env bash\nexec bash -l -c -- \"$*\"\n' > /usr/local/bin/entrypoint && \
    chmod +x /usr/local/bin/entrypoint

ENTRYPOINT ["/usr/local/bin/entrypoint"]
