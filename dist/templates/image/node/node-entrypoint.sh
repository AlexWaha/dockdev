#!/bin/sh

cd /var/www/html

if [ -f package.json ]; then
  echo "[Node Entrypoint] package.json found, installing..."
  npm install

  if [ -f package-lock.json ]; then
    echo "[Node Entrypoint] Running build..."
    npm run build
  else
    echo "[Node Entrypoint] No package-lock.json, skipping build."
  fi
else
  echo "[Node Entrypoint] No package.json found. Skipping."
fi

exec "$@"