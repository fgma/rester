#!/bin/sh

for REPO_NAME in repo1 repo2 repo3; do
    REPO_PATH=/tmp/rester_test/$REPO_NAME
    mkdir -p $REPO_PATH
    export RESTIC_PASSWORD=$REPO_NAME
    restic init -r $REPO_PATH
done
