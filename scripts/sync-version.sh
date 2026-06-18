#!/usr/bin/env sh
# Sync root VERSION into frontend/package.json (and package-lock.json when present).
set -eu

ROOT="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
VERSION="$(tr -d '[:space:]' < "$ROOT/VERSION")"

node -e "
const fs = require('fs');
const path = require('path');
const version = process.argv[1];
const pkgPath = path.join('$ROOT', 'frontend/package.json');
const pkg = JSON.parse(fs.readFileSync(pkgPath, 'utf8'));
pkg.version = version;
fs.writeFileSync(pkgPath, JSON.stringify(pkg, null, 2) + '\n');

const lockPath = path.join('$ROOT', 'frontend/package-lock.json');
if (fs.existsSync(lockPath)) {
  const lock = JSON.parse(fs.readFileSync(lockPath, 'utf8'));
  lock.version = version;
  if (lock.packages && lock.packages['']) lock.packages[''].version = version;
  fs.writeFileSync(lockPath, JSON.stringify(lock, null, 2) + '\n');
}
const composePath = path.join('$ROOT', 'docker-compose.yml');
if (fs.existsSync(composePath)) {
  let compose = fs.readFileSync(composePath, 'utf8');
  compose = compose.replace(
    /ECHOSTATE_VERSION:-0\\.0\\.1-beta\\.\\d+/g,
    'ECHOSTATE_VERSION:-' + version
  );
  fs.writeFileSync(composePath, compose);
}

const envExamplePath = path.join('$ROOT', '.env.example');
if (fs.existsSync(envExamplePath)) {
  let envExample = fs.readFileSync(envExamplePath, 'utf8');
  if (/^ECHOSTATE_VERSION=/m.test(envExample)) {
    envExample = envExample.replace(/^ECHOSTATE_VERSION=.*$/m, 'ECHOSTATE_VERSION=' + version);
  } else {
    envExample = 'ECHOSTATE_VERSION=' + version + '\\n' + envExample;
  }
  fs.writeFileSync(envExamplePath, envExample);
}

console.log('Synced version', version);
" "$VERSION"