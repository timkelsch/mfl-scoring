#!/bin/bash

set -euo pipefail

AWS_PROFILE="sam-personal-dev"
SAM=$(which sam)

$(which nvm) unload
$SAM build
$SAM deploy
