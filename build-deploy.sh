#!/bin/bash

set -xeuo pipefail

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm

nvm unload

export AWS_PROFILE="sam-personal-dev"
SAM=$(which sam)

$SAM build
$SAM deploy
