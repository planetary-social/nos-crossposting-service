#!/bin/bash
set -e

cd ./frontend/nos-crossposting-service-frontend

echo "Running yarn build"
rm -rf dist
yarn build

cd ../../

echo "Copying frontend files"
cd ./service/ports/http/frontend
rm -rf ./img ./css ./js
cp -r ../../../../frontend/nos-crossposting-service-frontend/dist/. ./
cd ../../../../