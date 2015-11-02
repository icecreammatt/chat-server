#!/bin/sh

openssl req -new -nodes -x509 -out server.pem -keyout server.key -days 365
