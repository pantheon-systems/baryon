#!/bin/bash

set -xe

# ensure we have autotag
if [ ! -d "$HOME/bin" ]; then
  mkdir -p ~/bin
fi

if [ ! -f "$HOME/bin/autotag" ]; then
  AUTOTAG_URL=$(curl -silent -o - -L https://api.github.com/repos/pantheon-systems/autotag/releases/latest | grep 'browser_' | cut -d\" -f4)
  # handle the off chance that this wont work with some pre-set version
  if [ -z "$AUTOTAG_URL" ] ;  then
    AUTOTAG_URL="https://github.com/pantheon-systems/autotag/releases/download/v0.0.3/autotag.linux.x86_64"
  fi
  curl -L $AUTOTAG_URL -o ~/bin/autotag
  chmod 755 ~/bin/autotag
fi
