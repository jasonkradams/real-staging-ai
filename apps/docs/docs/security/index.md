# Security

Security documentation and best practices for Real Staging AI.

## Overview

Real Staging AI implements multiple layers of security to protect user data, ensure service integrity, and prevent unauthorized access.

## Key Security Features

### Authentication & Authorization

- **OAuth 2.0 / OIDC** via Auth0
- **JWT tokens** with RS256 signing
- **Token validation** on every request
- **User isolation** via database foreign keys

[Learn more →](../guides/authentication.md)

### Data Protection

**In Transit:**
- TLS/HTTPS for all external communication
- Secure websocket connections for SSE

**At Rest:**
- S3 server-side encryption
- Encrypted database connections
- Secure credential storage

### API Security

**Rate Limiting:**
- 100 requests/minute per user
- 1000 requests/hour per user
- DDoS protection

**Input Validation:**
- Strict schema validation
- Content-Type enforcement
- File size limits
- Sanitization of user inputs

**CORS:**
- Configured allowed origins
- Preflight request handling
- Credential support

### Webhook Security

Stripe webhooks are secured with:
- HMAC signature verification
- Timestamp validation
- Idempotency checks
- Event replay prevention

[Read the detailed guide →](stripe-webhooks.md)

## Security Best Practices

### For Developers

✅ **DO:**
- Use environment variables for secrets
- Validate all inputs
- Implement proper error handling (don't leak information)
- Use prepared statements for SQL queries
- Log security events
- Keep dependencies updated
- Follow principle of least privilege

❌ **DON'T:**
- Hardcode credentials
- Commit secrets to version control
- Trust client-side validation alone
- Expose internal error details
- Use weak cryptographic algorithms
- Disable security features in production

### For Operators

✅ **DO:**
- Rotate secrets regularly
- Monitor for suspicious activity
- Keep systems patched
- Use secure configuration management
- Implement backup and recovery
- Enable audit logging
- Use network segmentation

❌ **DON'T:**
- Use default passwords
- Expose admin interfaces publicly
- Share credentials between environments
- Ignore security alerts
- Skip security updates

## Secure Configuration

### Required Security Settings

**Production Environment:**
```yaml
# config/prod.yml
server:
  tls_enabled: true
  tls_cert_file: /etc/certs/tls.crt
  tls_key_file: /etc/certs/tls.key

auth0:
  domain: "production.auth0.com"
  audience: "https://api.real-staging.ai"

stripe:
  webhook_secret: ${STRIPE_WEBHOOK_SECRET}

database:
  ssl_mode: "require"
```

### Secrets Management

**Never commit:**
- API keys
- Database passwords
- Webhook secrets
- TLS private keys
- OAuth client secrets

**Use instead:**
- Environment variables
- Secrets management tools (AWS Secrets Manager, HashiCorp Vault)
- Encrypted configuration files
- Kubernetes secrets

## Threat Model

### Potential Threats

| Threat | Mitigation |
|--------|-----------|
| **Unauthorized API access** | JWT validation, rate limiting |
| **SQL injection** | Prepared statements, input validation |
| **XSS attacks** | Output encoding, CSP headers |
| **CSRF** | SameSite cookies, CSRF tokens |
| **DDoS** | Rate limiting, WAF, CDN |
| **Data breaches** | Encryption, access controls, monitoring |
| **Webhook spoofing** | Signature verification |
| **Token theft** | Short expiration, HTTPS only |

## Incident Response

### If You Discover a Security Issue

**DO NOT** open a public GitHub issue.

Instead:
1. Email security@real-staging.ai with details
2. Include steps to reproduce (if applicable)
3. Wait for acknowledgment before disclosure
4. Allow time for fix and deployment

### Security Update Process

1. **Report received** - Acknowledged within 24 hours
2. **Assessment** - Severity determined
3. **Fix developed** - Patch created and tested
4. **Deployment** - Rolled out to production
5. **Disclosure** - Public disclosure after fix deployed

## Compliance

### Data Privacy

- **GDPR compliant** - European users
- **CCPA compliant** - California users
- **Data retention policies** - Configurable retention
- **Right to deletion** - User data can be removed

### Industry Standards

- OAuth 2.0 / OpenID Connect
- JWT (RFC 7519)
- TLS 1.3
- OWASP Top 10 mitigations

## Security Auditing

### Logging

Security events are logged:
```json
{
  "timestamp": "2025-10-12T20:30:00Z",
  "level": "warn",
  "event": "authentication_failure",
  "ip": "203.0.113.45",
  "user_agent": "...",
  "reason": "invalid_token"
}
```

### Monitoring

Key security metrics:
- Failed authentication attempts
- Rate limit violations
- Invalid webhook signatures
- Unusual API usage patterns
- Database connection failures

### Alerts

Alerts are triggered for:
- High rate of authentication failures
- Webhook signature verification failures
- Unusual geographic access patterns
- API error rate spikes

## Regular Security Tasks

### Weekly
- [ ] Review security logs
- [ ] Check for failed login patterns
- [ ] Monitor rate limit violations

### Monthly
- [ ] Update dependencies
- [ ] Review access controls
- [ ] Audit API tokens
- [ ] Check TLS certificate expiration

### Quarterly
- [ ] Rotate secrets
- [ ] Review security policies
- [ ] Conduct security training
- [ ] Penetration testing (if applicable)

## Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Auth0 Security Best Practices](https://auth0.com/docs/security)
- [Stripe Security](https://stripe.com/docs/security)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)

---

**Related Documentation:**
- [Authentication Guide](../guides/authentication.md)
- [Stripe Webhooks](stripe-webhooks.md)
- [Configuration](../guides/configuration.md)
