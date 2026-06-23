---
title: Sample Configuration
description: Configure the sample component
---

{{< docs/shared source="alloy" lookup="agent-deprecation.md" version="next" >}}

<!-- Navigation placeholder -->

# Sample Configuration

This is the main configuration page.

## Authentication

Configure authentication for the sample component.

```yaml
auth:
  enabled: true
  type: bearer
```

Use bearer tokens for secure access.

## Storage

Storage settings control where data is persisted.

### Local storage

Local storage uses the filesystem.

### Remote storage

Remote storage supports S3 and GCS.

## Advanced

Advanced settings for power users.
