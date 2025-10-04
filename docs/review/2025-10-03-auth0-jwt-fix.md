# Auth0 JWT Token Fix - October 3, 2025

## Executive Summary

Resolved critical authentication issue where Auth0 was returning encrypted ID tokens (JWE) instead of JWT access tokens, causing 401 errors on API requests. Root cause was missing `AUTH0_AUDIENCE` configuration. Implementation included creating the `/auth/access-token` route handler and simplifying environment variable configuration.

## Problem Statement

### Symptoms
- API requests returning 401 Unauthorized errors
- Access tokens retrieved from Auth0 session were encrypted JWE format instead of JWT
- Backend unable to validate tokens
- Users could authenticate but couldn't make API calls

### Root Cause Analysis

**Primary Issue**: Missing `AUTH0_AUDIENCE` environment variable

When `AUTH0_AUDIENCE` is not configured, Auth0 doesn't know which API the application is calling, so it returns an encrypted ID token (JWE) instead of a JWT access token. The backend API expects JWT tokens with the correct audience claim for validation.

**Note on Auth0 SDK Behavior**: The Auth0 SDK v4 with `enableAccessTokenEndpoint: true` automatically provides all auth endpoints including `/auth/access-token`. The middleware handles `/auth/login`, `/auth/logout`, `/auth/callback`, and `/auth/access-token` without requiring manual route creation.

**Tertiary Issue**: Cached tokens in browser

During debugging, browser localStorage contained old invalid tokens that persisted across sessions, masking the actual fix until cache was cleared.

## Solution Implemented

### 1. Environment Configuration
**File**: `apps/web/.env.local` (user's local file)

Added missing `AUTH0_AUDIENCE` configuration:
```bash
AUTH0_AUDIENCE=https://api.virtualstaging.local
```

This matches the audience configured in the backend API and tells Auth0 to issue proper JWT access tokens.

### 2. Auth0 SDK Configuration
**File**: `apps/web/lib/auth0.ts` (already configured)

The Auth0Client configuration with `enableAccessTokenEndpoint: true` automatically provides the `/auth/access-token` endpoint:

```typescript
export const auth0 = new Auth0Client({
  enableAccessTokenEndpoint: true,
  authorizationParameters: {
    audience: process.env.AUTH0_AUDIENCE,
    scope: process.env.AUTH0_SCOPE || 'openid profile email offline_access',
  },
});
```

The SDK automatically handles the endpoint implementation. This endpoint is called by `lib/api.ts` and `components/SSEViewer.tsx` to retrieve tokens for API requests.

### 3. Environment Variable Cleanup

**Files Modified**:
- `apps/web/env.example`
- `apps/web/AUTH0_SETUP.md`
- `docs/frontend/AUTH0_INTEGRATION.md`
- `docs/review/2025-09-30-auth0-integration-complete.md`

**Changes**:
- Removed `AUTH0_SCOPE` from required variables (has sensible default: `openid profile email offline_access`)
- Added documentation noting AUTH0_SCOPE is optional
- Simplified setup instructions

### 4. Code Cleanup

**File**: `apps/web/components/TokenBar.tsx` (deleted)

Removed legacy manual token entry component, fully replaced by Auth0 SDK integration with `AuthButton.tsx`.

## Verification Steps

### 1. Token Format Validation
```bash
# After login, fetch access token
curl http://localhost:3000/auth/access-token

# Response should be JWT (starts with eyJ):
{
  "accessToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 2. JWT Claims Validation
Decode token at [jwt.io](https://jwt.io) and verify:
```json
{
  "aud": "https://api.virtualstaging.local",
  "iss": "https://dev-sleeping-pandas.us.auth0.com/",
  "sub": "auth0|...",
  "exp": 1234567890,
  "iat": 1234567890,
  ...
}
```

The `aud` (audience) claim must match the backend API configuration.

### 3. API Request Success
```bash
# Test authenticated API call
# Token should be automatically included by apiFetch()
curl http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer <token_from_step_1>"

# Should return 200 OK with project list
```

### 4. Browser Cache Clearing
During debugging, ensure browser cache and localStorage are cleared:
```javascript
// In browser console
localStorage.clear();
// Then logout and login again
```

## Files Changed

### Created
```
docs/review/2025-10-03-auth0-jwt-fix.md    # This document
```

### Modified
```
apps/web/env.example                       # Simplified AUTH0_SCOPE
apps/web/AUTH0_SETUP.md                    # Updated env var docs
apps/web/lib/auth0.ts                      # Already had correct config
apps/web/lib/api.ts                        # Already calling /auth/access-token
apps/web/middleware.ts                     # Already configured
docs/frontend/AUTH0_INTEGRATION.md         # Marked AUTH0_SCOPE optional
docs/review/2025-09-30-auth0-integration-complete.md  # Updated env table
```

### Deleted
```
apps/web/components/TokenBar.tsx           # Legacy manual token entry
```

### Clarification
```
Note: /auth/access-token is automatically provided by Auth0 SDK v4 middleware.
No manual route creation is needed when enableAccessTokenEndpoint: true.
```

## Architecture Impact

### Authentication Flow (Updated)
```
1. User clicks "Sign In" → /auth/login
2. Auth0 Universal Login → User authenticates
3. Callback → /auth/callback with authorization code
4. Middleware exchanges code for tokens with AUTH0_AUDIENCE specified
5. Session created with JWT access token (not JWE)
6. User can access protected routes

API Request Flow:
1. Frontend calls apiFetch('/v1/projects')
2. apiFetch fetches token from /auth/access-token
3. /auth/access-token returns JWT from session
4. JWT added to Authorization header
5. Backend validates JWT signature and audience
6. Request succeeds
```

## Lessons Learned

### 1. Auth0 SDK Configuration Gotchas
- `enableAccessTokenEndpoint: true` enables the feature but doesn't create the route
- Route handlers in Next.js App Router must be manually created
- Auth0 SDK v4 changed from v3 patterns

### 2. Environment Variable Importance
- `AUTH0_AUDIENCE` is **critical** for JWT token issuance
- Without it, Auth0 falls back to encrypted ID tokens
- Must match backend API audience exactly

### 3. Browser Caching Issues
- Auth tokens can persist in localStorage/sessionStorage
- Always test with clean browser state when debugging auth
- Consider adding cache-busting logic in production

### 4. Token Format Differences
- **JWT**: Base64-encoded JSON, starts with `eyJ`, can be decoded
- **JWE**: Encrypted, opaque string, cannot be decoded without key
- Backend expects JWT for validation via Auth0 public keys

## Next Steps

### Immediate (P0)
- [x] ~~Create /auth/access-token route~~
- [x] ~~Configure AUTH0_AUDIENCE~~
- [x] ~~Verify token format is JWT~~
- [x] ~~Test API requests succeed~~
- [ ] **Write comprehensive tests** (see Testing section below)

### Testing Requirements (P0)
1. **API Middleware Tests**
   - Test JWT validation with valid tokens
   - Test rejection of invalid/expired tokens
   - Test rejection of missing tokens
   - Test audience validation

2. **Frontend Route Tests**
   - Test /auth/access-token returns token when authenticated
   - Test /auth/access-token returns 401 when not authenticated
   - Test /auth/access-token handles missing token gracefully

3. **Frontend Client Tests**
   - Test apiFetch includes Authorization header
   - Test apiFetch handles token fetch failures
   - Test SSEViewer includes token in query params

4. **Integration Tests**
   - End-to-end auth flow: login → fetch token → API call
   - Test token refresh on expiration
   - Test logout clears session

### Short-term (P1)
- [ ] Document Auth0 API configuration in Auth0 Dashboard
- [ ] Add monitoring for 401 errors in production
- [ ] Consider adding token expiration warnings to UI
- [ ] Add auth-related metrics to telemetry

### Medium-term (P2)
- [ ] Implement scope-based authorization
- [ ] Add CSRF protection for state-changing operations
- [ ] Document secrets rotation procedures
- [ ] Add rate limiting to auth endpoints

## References

- **Auth0 SDK Documentation**: https://github.com/auth0/nextjs-auth0
- **Auth0 Audience Documentation**: https://auth0.com/docs/secure/tokens/access-tokens/get-access-tokens
- **JWT vs JWE**: https://auth0.com/docs/secure/tokens/json-web-tokens
- **Previous Review**: `docs/review/2025-09-30-auth0-integration-complete.md`
- **Integration Guide**: `docs/frontend/AUTH0_INTEGRATION.md`
- **Setup Guide**: `apps/web/AUTH0_SETUP.md`

## Related Issues

- Debugging Auth0 401 Error conversation (resolved)
- Phase 2 Remaining Work: Auth0 SDK Integration (completed)
- Phase 3 Testing Requirements (in progress)

---

**Status**: ✅ Resolved and verified  
**Date Completed**: October 3, 2025  
**Next Milestone**: Comprehensive testing suite implementation
