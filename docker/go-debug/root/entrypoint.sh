#!/bin/bash

if [[ -z "${DEPLOY_ENV}" ]]; then
  echo "RUNNING IN PRODUCTION MODE"
else
  GO_WORK_DIR=${GO_WORK_DIR:-$GOPATH/src}
  cd "${GO_WORK_DIR}"

  go mod download
  go get -d
fi

exec "$@"
