#!/bin/bash
set -e

git pull
git add .
git commit -m "."
git push
