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
# Enable authentication for all API endpoints.
# Supported types: bearer, basic, api-key.
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

## Templates

Hugo template examples for embedding docs.

```html
<!-- This comment should be preserved inside the code block -->
<div class="docs-content">
  {{< docs/shared source="tempo" lookup="config.md" >}}
</div>
```

## Advanced

Advanced settings for power users.
