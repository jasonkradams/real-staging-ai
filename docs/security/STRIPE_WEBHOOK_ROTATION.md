# Stripe Webhook Secret Rotation

This document describes the procedure for rotating the Stripe webhook signing secret (`STRIPE_WEBHOOK_SECRET`) to maintain security and compliance.

## When to Rotate

Rotate the webhook secret in the following scenarios:

1. **Scheduled Rotation**: Every 90 days as part of routine security maintenance
2. **Security Incident**: Immediately if the secret is compromised or exposed
3. **Team Changes**: When team members with access to production secrets leave
4. **Compliance Requirements**: Per PCI-DSS or other applicable regulations

## Prerequisites

- Admin access to Stripe Dashboard
- Access to production environment variables (AWS Secrets Manager, Kubernetes secrets, or deployment platform)
- Access to deployment pipeline (CI/CD)
- Rollback plan prepared

## Rotation Procedure

### Phase 1: Preparation (5-10 minutes)

1. **Schedule Maintenance Window** (Optional but Recommended)
   ```bash
   # For zero-downtime rotation, this is optional
   # For critical systems, schedule a brief maintenance window
   ```

2. **Backup Current Configuration**
   ```bash
   # Document current webhook endpoint and secret
   echo "Current STRIPE_WEBHOOK_SECRET: stripe_webhook_secret_****" > rotation_backup.txt
   echo "Date: $(date)" >> rotation_backup.txt
   ```

3. **Verify Monitoring**
   - Ensure Stripe webhook logs are accessible
   - Set up alerts for webhook signature validation failures
   - Verify application error logging is working

### Phase 2: Create New Secret (2-3 minutes)

1. **Log into Stripe Dashboard**
   - Navigate to: Developers → Webhooks
   - Select your webhook endpoint (e.g., `https://api.yourdomain.com/api/v1/stripe/webhook`)

2. **Roll the Signing Secret**
   - Click "Roll secret" or create a new endpoint
   - **Important**: Stripe will show the new secret only once
   - Copy the new secret immediately: `whsec_...`

3. **Document the New Secret**
   ```bash
   # Store in secure password manager
   # DO NOT commit to git
   echo "New STRIPE_WEBHOOK_SECRET: whsec_..." >> rotation_backup.txt
   ```

### Phase 3: Update Application (5-10 minutes)

#### Option A: Zero-Downtime Rotation (Recommended)

For applications that support multiple webhook secrets (requires code changes):

```go
// Example: Support validating against multiple secrets during rotation
func (h *DefaultWebhookHandler) ValidateWebhookSignature(payload []byte, signature string) error {
    secrets := []string{
        os.Getenv("STRIPE_WEBHOOK_SECRET"),
        os.Getenv("STRIPE_WEBHOOK_SECRET_OLD"), // Fallback during rotation
    }
    
    for _, secret := range secrets {
        if secret == "" {
            continue
        }
        event, err := webhook.ConstructEvent(payload, signature, secret)
        if err == nil {
            return nil
        }
    }
    
    return errors.New("webhook signature validation failed")
}
```

1. Deploy application with old secret still active
2. Set `STRIPE_WEBHOOK_SECRET_OLD=<current_secret>`
3. Set `STRIPE_WEBHOOK_SECRET=<new_secret>`
4. Deploy and verify webhooks work
5. After 24-48 hours, remove `STRIPE_WEBHOOK_SECRET_OLD`

#### Option B: Brief Downtime Rotation

1. **Update Environment Variable**
   
   AWS Secrets Manager:
   ```bash
   aws secretsmanager update-secret \
     --secret-id prod/stripe/webhook-secret \
     --secret-string "whsec_NEW_SECRET_HERE"
   ```

   Kubernetes:
   ```bash
   kubectl create secret generic stripe-webhook-secret \
     --from-literal=STRIPE_WEBHOOK_SECRET=whsec_NEW_SECRET_HERE \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

   Docker Compose / Environment File:
   ```bash
   # Update .env or docker-compose.yml
   STRIPE_WEBHOOK_SECRET=whsec_NEW_SECRET_HERE
   ```

2. **Deploy/Restart Application**
   ```bash
   # Kubernetes
   kubectl rollout restart deployment/api
   
   # Docker Compose
   docker compose restart api
   
   # Systemd
   sudo systemctl restart virtual-staging-api
   ```

3. **Verify Deployment**
   ```bash
   # Check application logs for startup
   kubectl logs -f deployment/api
   
   # Verify health endpoint
   curl https://api.yourdomain.com/health
   ```

### Phase 4: Verification (10-15 minutes)

1. **Test Webhook Endpoint**
   
   Use Stripe CLI to send test events:
   ```bash
   stripe trigger checkout.session.completed --api-key=sk_test_...
   ```

   Or use Stripe Dashboard:
   - Go to Developers → Webhooks
   - Click "Send test webhook"
   - Select event type (e.g., `checkout.session.completed`)
   - Click "Send test webhook"

2. **Check Application Logs**
   ```bash
   # Look for successful webhook processing
   kubectl logs -f deployment/api | grep "webhook"
   
   # Should see:
   # {"level":"INFO","msg":"webhook received","event_type":"checkout.session.completed"}
   # {"level":"INFO","msg":"webhook processed successfully"}
   ```

3. **Monitor for Errors**
   ```bash
   # Check for signature validation errors
   kubectl logs deployment/api | grep "signature"
   
   # Should NOT see:
   # {"level":"ERROR","msg":"webhook signature validation failed"}
   ```

4. **Verify in Stripe Dashboard**
   - Go to Developers → Webhooks → Your Endpoint
   - Check "Recent Deliveries" tab
   - Confirm events show "Success" status (200 response)

### Phase 5: Cleanup (2-3 minutes)

1. **Remove Old Secret from Stripe** (After 24-48 hours)
   - If you created a new endpoint, delete the old one
   - If you rolled the secret, the old one is automatically invalidated

2. **Update Documentation**
   ```bash
   # Update runbook with rotation date
   echo "Last rotated: $(date)" >> docs/security/STRIPE_WEBHOOK_ROTATION.md
   echo "Rotated by: $(whoami)" >> rotation_log.txt
   ```

3. **Secure Disposal**
   ```bash
   # Securely delete backup file
   shred -u rotation_backup.txt
   
   # Clear shell history if secret was in commands
   history -c
   ```

## Rollback Procedure

If the new secret causes issues:

1. **Immediate Rollback**
   ```bash
   # Revert to old secret
   kubectl set env deployment/api STRIPE_WEBHOOK_SECRET=<OLD_SECRET>
   
   # Or update secrets manager and redeploy
   aws secretsmanager update-secret \
     --secret-id prod/stripe/webhook-secret \
     --secret-string "<OLD_SECRET>"
   ```

2. **In Stripe Dashboard**
   - If you created a new endpoint, point your app back to the old endpoint
   - Roll the secret again to regenerate the old one if needed

3. **Verify Rollback**
   - Test webhook delivery
   - Check application logs
   - Monitor error rates

## Automation Considerations

For frequent rotations or large deployments, consider:

1. **Automated Rotation Script**
   ```bash
   #!/bin/bash
   # rotate_stripe_webhook.sh
   
   # Fetch new secret from Stripe API
   NEW_SECRET=$(stripe webhooks update we_xxx --secret-refresh)
   
   # Update secrets manager
   aws secretsmanager update-secret \
     --secret-id prod/stripe/webhook-secret \
     --secret-string "$NEW_SECRET"
   
   # Trigger deployment
   kubectl rollout restart deployment/api
   ```

2. **Secret Rotation Lambda/Cron**
   - Schedule automatic rotation every 90 days
   - Send notifications to ops team
   - Verify webhooks after rotation

3. **Multiple Webhook Endpoints**
   - Use separate webhook endpoints for different environments
   - Rotate production and staging independently
   - Test rotation procedure in staging first

## Security Best Practices

1. **Secret Storage**
   - ✅ Store in AWS Secrets Manager, Google Secret Manager, or HashiCorp Vault
   - ✅ Use environment variables, never hardcode
   - ❌ Never commit secrets to version control
   - ❌ Never log the full secret value

2. **Access Control**
   - Limit who can access production secrets
   - Use IAM roles and least-privilege principles
   - Audit secret access logs regularly

3. **Monitoring**
   - Set up alerts for webhook signature validation failures
   - Monitor webhook delivery success rates in Stripe Dashboard
   - Track secret rotation dates in compliance tracking tool

4. **Disaster Recovery**
   - Document all webhook endpoints and their purposes
   - Maintain secure backup of current configurations
   - Test rollback procedure in staging environment

## Troubleshooting

### Webhook Signature Validation Failures

**Symptoms**: 400 errors in Stripe Dashboard, "signature validation failed" in logs

**Causes**:
- New secret not yet deployed
- Application cached old secret
- Multiple application instances with different secrets

**Solutions**:
1. Verify all application instances have new secret
2. Restart all application pods/containers
3. Check for typos in secret value
4. Verify no trailing whitespace in secret

### Webhook Endpoint Not Receiving Events

**Symptoms**: No webhook events in application logs

**Causes**:
- Stripe endpoint disabled
- Firewall blocking Stripe IPs
- Application health check failing

**Solutions**:
1. Check webhook endpoint status in Stripe Dashboard
2. Verify endpoint URL is correct and accessible
3. Check application logs for startup errors
4. Test with `stripe trigger` CLI command

### Partial Deployment Issues

**Symptoms**: Some webhooks succeed, others fail

**Causes**:
- Rolling deployment in progress
- Some instances have old secret

**Solutions**:
1. Wait for deployment to complete
2. Force restart all instances
3. Use blue-green deployment for zero-downtime

## Compliance & Audit Trail

Document each rotation:

```
Date: 2025-09-29
Rotated By: ops-team
Reason: Scheduled 90-day rotation
Old Secret Last 4: ...abc1
New Secret Last 4: ...def2
Downtime: 0 minutes
Issues: None
Verification: All test webhooks successful
```

Store audit trail in:
- Compliance tracking system
- Security incident log
- Change management system

## References

- [Stripe Webhook Signatures Documentation](https://stripe.com/docs/webhooks/signatures)
- [Stripe Security Best Practices](https://stripe.com/docs/security/best-practices)
- [PCI-DSS Key Rotation Requirements](https://www.pcisecuritystandards.org/)
- Internal: `docs/operations/DEPLOYMENT.md`
- Internal: `docs/security/SECRETS_MANAGEMENT.md`

## Last Rotation

```
Date: [To be filled during rotation]
Rotated By: [Team member name]
Next Scheduled: [Date + 90 days]
```
