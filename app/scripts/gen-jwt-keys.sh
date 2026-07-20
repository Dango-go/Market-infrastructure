#!/usr/bin/env bash
# Generates an RSA-2048 keypair for the auth service to sign access tokens (RS256).
# The private key is consumed by the auth service (JWT_PRIVATE_KEY_FILE); the public key
# is exposed by the service at /.well-known/jwks.json for other services to verify tokens.
set -euo pipefail

OUT_DIR="${1:-./secrets}"
mkdir -p "$OUT_DIR"

PRIV="$OUT_DIR/jwt_private.pem"
PUB="$OUT_DIR/jwt_public.pem"

if [[ -f "$PRIV" ]]; then
  echo "refusing to overwrite existing key: $PRIV" >&2
  exit 1
fi

openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out "$PRIV"
openssl rsa -in "$PRIV" -pubout -out "$PUB"

chmod 600 "$PRIV"
echo "wrote:"
echo "  private key -> $PRIV  (set JWT_PRIVATE_KEY_FILE to this path)"
echo "  public key  -> $PUB   (served via JWKS; share with other services)"
