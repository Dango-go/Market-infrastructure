Place JWT keys here for local docker-compose runs:

- jwt_private.pem
- jwt_public.pem

Expected paths inside containers:
- /workspace/backend/secrets/jwt_private.pem
- /workspace/backend/secrets/jwt_public.pem

You can generate them with the auth service helper script if present in the repo.
