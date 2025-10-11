# Replicate API Setup for Real Staging

This guide explains how to set up Replicate API integration for AI-powered virtual staging.

## Overview

Real Staging AI uses [Replicate](https://replicate.com) to run the actual AI model for staging images. The integration uses the `qwen/qwen-image-edit` model, which provides:

- **Fast generation**: ~9 seconds per image
- **Cost-effective**: ~$0.011 per image (~90 images per $1)
- **High quality**: Photorealistic SDXL-based outputs with ControlNet
- **No infrastructure**: Fully managed GPU inference

## Get Your API Token

### 1. Sign up for Replicate

Visit [replicate.com](https://replicate.com) and create a free account.

### 2. Get your API token

1. Go to your [account settings](https://replicate.com/account/api-tokens)
2. Create a new API token
3. Copy the token (starts with `r8_`)

### 3. Set the environment variable

#### For local development

Add to your `.env` file:

```bash
REPLICATE_API_TOKEN=r8_your_token_here
```

Or export it in your shell:

```bash
export REPLICATE_API_TOKEN=r8_your_token_here
```

#### For docker-compose

The `docker-compose.yml` already reads `REPLICATE_API_TOKEN` from your environment:

```bash
export REPLICATE_API_TOKEN=r8_your_token_here
make up
```

#### For production

Set as an environment variable in your deployment platform:

- **AWS**: Parameter Store or Secrets Manager
- **GCP**: Secret Manager
- **Kubernetes**: Secrets
- **Heroku/Render**: Environment variables in dashboard

## Model Configuration

### Default Model

The worker uses `qwen/qwen-image-edit` by default (configured in `docker-compose.yml`).

### Custom Model (Optional)

To use a different model, set:

```bash
REPLICATE_MODEL_VERSION=owner/model-name:version-hash
```

Examples:

- `rocketdigitalai/interior-design-sdxl` (slower but higher quality)
- `stability-ai/sdxl` (general purpose SDXL)
- Your own fine-tuned model

## Cost Management

### Pricing

- **Per-image cost**: ~$0.011
- **Monthly estimate**:
  - 100 images: ~$1.10
  - 1,000 images: ~$11
  - 10,000 images: ~$110

### Cost Controls

1. **Replicate Dashboard**: Monitor usage at [replicate.com/account/billing](https://replicate.com/account/billing)

2. **Set billing alerts**: Configure alerts in your Replicate account settings

3. **Rate limiting**: Consider adding rate limits in your application if needed

4. **Development**: Use a separate Replicate account for development/testing

## Configuration Reference

### Required Environment Variables

| Variable              | Description                         | Example           |
| --------------------- | ----------------------------------- | ----------------- |
| `REPLICATE_API_TOKEN` | Your Replicate API token            | `r8_abc123...`    |
| `S3_BUCKET_NAME`      | S3 bucket for storing staged images | `real-staging` |

### Optional Environment Variables

| Variable                  | Description              | Default                                          |
| ------------------------- | ------------------------ | ------------------------------------------------ |
| `REPLICATE_MODEL_VERSION` | Model to use for staging | `qwen/qwen-image-edit` |

## Testing the Integration

### 1. Start the services

```bash
export REPLICATE_API_TOKEN=r8_your_token_here
make up
```

### 2. Upload an image

Navigate to http://localhost:3000/upload and:

1. Create a project
2. Upload an interior image
3. Optionally select room type and style
4. Submit

### 3. Watch the processing

Go to http://localhost:3000/images to see:

- Status updates via SSE (queued → processing → ready)
- The staged image when complete

### 4. Check logs

```bash
docker-compose logs -f worker
```

Look for:

- `Replicate client initialized`
- `Processing stage job for image...`
- `Successfully staged image...`

## Troubleshooting

### "REPLICATE_API_TOKEN environment variable is required"

**Solution**: Export the token before starting:

```bash
export REPLICATE_API_TOKEN=r8_your_token_here
make up
```

### "Failed to create Replicate client"

**Causes**:

- Invalid token format
- Token has been revoked
- Network connectivity issues

**Solution**:

1. Verify token at [replicate.com/account/api-tokens](https://replicate.com/account/api-tokens)
2. Generate a new token if needed
3. Check network/firewall settings

### "Prediction timed out after 5 minutes"

**Causes**:

- Large image size
- Model overloaded
- Network issues

**Solutions**:

1. Reduce image size (max 10MB)
2. Retry the job
3. Check Replicate status at [status.replicate.com](https://status.replicate.com)

### Images not being staged

**Check**:

1. Worker logs: `docker-compose logs -f worker`
2. Replicate dashboard: [replicate.com/account](https://replicate.com/account)
3. S3 bucket: Verify images are being uploaded

### High costs

**Actions**:

1. Review usage in Replicate dashboard
2. Check for stuck/retrying jobs
3. Implement rate limiting if needed
4. Consider caching staged images

## Alternative: Self-Hosted

If you prefer to run the model yourself:

1. **GPU Requirements**: NVIDIA GPU with 24GB+ VRAM
2. **Framework**: ComfyUI or Diffusers
3. **Model**: RealVisXL V5.0-Lightning + ControlNet
4. **Cost**: Lower at scale (>10k images/month) but requires DevOps

See `docs/SELF_HOSTED_STAGING.md` for details (coming soon).

## Support

- **Replicate Docs**: https://replicate.com/docs
- **Model Page**: https://replicate.com/qwen/qwen-image-edit
- **Community**: https://discord.gg/replicate

---

**Pro Tip**: Start with Replicate for fast time-to-market. Consider self-hosting only if you're processing >10,000 images/month and have GPU infrastructure expertise.
