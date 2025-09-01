# Auth

## Auth0 Setup

- Create **API** in Auth0 → set Identifier (Audience) to match `JWT_AUDIENCE`
- Create **SPA** app for the web; configure callback/logout URLs
- Add RBAC (optional) and scopes if needed later

## API Verification (Echo Middleware)

- Fetch JWKS from `https://YOUR_DOMAIN.auth0.com/.well-known/jwks.json`
- Cache keys; validate issuer & audience; check `exp`, `nbf`.
- Inject `userID` from `sub` → map to local `users` row.

## Testing Auth

- Unit tests stub JWKS and sign tokens locally
- Integration tests can run without Auth by enabling a dev-only bypass flag
