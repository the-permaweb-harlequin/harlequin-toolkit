# 🌐 Install Script Hosting Setup

## Overview

The Harlequin CLI install script requires hosting at `install_cli_harlequin.daemongate.io` with specific endpoints for version management and binary distribution.

## 📋 Required Endpoints

### 1. Install Script
```
GET https://install_cli_harlequin.daemongate.io
Content-Type: text/plain
Returns: install_cli.sh script content
```

### 2. Releases API
```
GET https://install_cli_harlequin.daemongate.io/releases
Content-Type: application/json
Returns: Array of available releases
```

### 3. Binary Downloads
```
GET https://install_cli_harlequin.daemongate.io/releases/{version}/{platform}/{arch}
Content-Type: application/octet-stream
Returns: Binary file for platform/arch
```

## 🚀 Implementation Options

### Option 1: Static Hosting + CDN
**Best for**: Simple setup, high availability

```bash
# File structure:
/
├── index.html                          # Redirects to install script
├── install_cli.sh                      # Install script
├── releases.json                       # Releases metadata
└── releases/
    ├── latest/
    │   ├── linux/
    │   │   ├── amd64                   # Binary file
    │   │   └── arm64                   # Binary file
    │   ├── darwin/
    │   │   ├── amd64                   # Binary file
    │   │   └── arm64                   # Binary file
    │   └── windows/
    │       └── amd64.exe               # Binary file
    └── 1.2.3/
        └── ... (same structure)
```

**Implementation:**
- AWS S3 + CloudFront
- Netlify/Vercel static hosting
- GitHub Pages (if public)

### Option 2: Simple API Server
**Best for**: Dynamic responses, usage analytics

```javascript
// Express.js example
const express = require('express');
const app = express();

// Serve install script
app.get('/', (req, res) => {
  res.set('Content-Type', 'text/plain');
  res.sendFile('/path/to/install_cli.sh');
});

// Releases API
app.get('/releases', (req, res) => {
  res.json([
    {
      "tag_name": "cli-v1.2.3",
      "version": "1.2.3",
      "assets": [
        {
          "name": "harlequin-linux-amd64",
          "url": "https://install_cli_harlequin.daemongate.io/releases/1.2.3/linux/amd64"
        }
        // ... more assets
      ]
    }
  ]);
});

// Binary downloads
app.get('/releases/:version/:platform/:arch', (req, res) => {
  const { version, platform, arch } = req.params;
  const filename = \`harlequin-\${platform}-\${arch}\${platform === 'windows' ? '.exe' : ''}\`;
  res.download(\`/binaries/\${version}/\${filename}\`);
});
```

### Option 3: Serverless Functions
**Best for**: Cost efficiency, auto-scaling

```javascript
// Vercel API route: api/releases.js
export default function handler(req, res) {
  if (req.method === 'GET') {
    res.json([
      {
        "tag_name": "cli-v1.2.3", 
        "version": "1.2.3"
      }
    ]);
  }
}

// Vercel API route: api/releases/[...params].js
export default function handler(req, res) {
  const [version, platform, arch] = req.query.params;
  // Stream binary from storage
}
```

## 🔧 Setup Steps

### 1. Domain Configuration
```bash
# Set up DNS for install_cli_harlequin.daemongate.io
# Point to your hosting provider
```

### 2. SSL Certificate
```bash
# Ensure HTTPS is enabled
# Use Let's Encrypt, Cloudflare, or provider SSL
```

### 3. GitHub Actions Integration
```yaml
# In .github/workflows/release.yml
- name: Upload binaries to hosting
  run: |
    # Upload to your chosen hosting solution
    aws s3 sync dist/cli/ s3://your-bucket/releases/${{ version }}/
    # OR
    curl -X POST "https://api.yourhost.com/upload" -F "file=@binary"
    # OR
    rsync -av dist/cli/ user@server:/var/www/releases/${{ version }}/
```

### 4. Releases API Updates
```bash
# Update releases.json when new version is released
# This can be done via:
# - Direct file update (static hosting)
# - Database update (dynamic API)
# - GitHub API integration
```

## 📊 Recommended Architecture

### For Production Use:
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CloudFlare    │    │     Vercel       │    │   AWS S3        │
│   (DNS + SSL)   │───▶│  (API + Script)  │───▶│   (Binaries)    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
        │                        │                        │
        ▼                        ▼                        ▼
install_cli_harlequin     /releases API           Binary Storage
.daemongate.io           /install_cli.sh          /releases/*/*
```

**Benefits:**
- ✅ **High availability** - CDN distribution
- ✅ **Fast downloads** - Edge caching
- ✅ **Cost effective** - Serverless scaling
- ✅ **Analytics** - Usage tracking
- ✅ **Security** - HTTPS + DDoS protection

## 🎯 Quick Start with Vercel

1. **Create Vercel project**
```bash
npm i -g vercel
vercel init harlequin-installer
```

2. **Project structure**
```
harlequin-installer/
├── public/
│   └── install_cli.sh
├── api/
│   ├── releases.js
│   └── releases/
│       └── [...params].js
└── vercel.json
```

3. **Configure routing**
```json
// vercel.json
{
  "routes": [
    { "src": "/", "dest": "/public/install_cli.sh" },
    { "src": "/releases", "dest": "/api/releases.js" },
    { "src": "/releases/(.*)", "dest": "/api/releases/$1.js" }
  ]
}
```

4. **Deploy**
```bash
vercel --prod
```

5. **Set custom domain**
```bash
vercel domains add install_cli_harlequin.daemongate.io
```

This setup provides a professional, scalable hosting solution for your install script! 🎭
