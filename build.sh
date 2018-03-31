HASH = `git rev-parse --verify HEAD`

sudo docker build -t evilben/travel_auth:$HASH auth/