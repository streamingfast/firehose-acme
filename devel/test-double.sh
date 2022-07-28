#!/usr/bin/env bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )"

main() {
  pushd "$ROOT" &> /dev/null

    while true; do
        echo "Running until hitting the error"
        printf "" > /tmp/double-close.log

        ./devel/standard/start.sh -c 2>&1 | tee /tmp/double-close.log

        if grep -q 'panic: close of closed channel' "/tmp/double-close.log"; then
            echo "Hit the error, finished"
            exit 0
        fi

        echo "Retrying to hit the error"
        echo ""
        echo ""
    done

    echo "Completed"
}

main "$@"