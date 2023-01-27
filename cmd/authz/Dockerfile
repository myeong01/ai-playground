FROM alpine:3.10
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

ENV USER=authservice
ENV GROUP=authservice

# Add new user to run as
RUN addgroup -S -g 111 $GROUP && adduser -S -G $GROUP $USER
ENV APP_HOME=/home/$USER
WORKDIR $APP_HOME

# Copy in binary and give permissions
COPY bin/oidc-authservice $APP_HOME
RUN chmod +x $APP_HOME/oidc-authservice
RUN chown -R $USER:$GROUP $APP_HOME

USER $USER

ENTRYPOINT [ "./oidc-authservice" ]

