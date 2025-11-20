# MCP (Model Context Protocol) Server Setup Guide

This guide walks you through setting up MCP servers for enhanced Claude Code capabilities when working with the Last-Mile Accelerator codebase.

## What is MCP?

Model Context Protocol (MCP) enables Claude Code to access external tools, databases, and APIs beyond the built-in capabilities. Think of MCP servers as plugins that extend Claude Code's functionality.

## Recommended MCP Servers for LMA

Based on your codebase analysis, these MCP servers provide the most value:

### 1. GitHub MCP (Essential)
**Why:** Manage 20+ open PRs, track issues, view workflows
**Capabilities:**
- View and create issues
- Manage pull requests
- Check CI/CD status
- Review workflow runs
- Search code across repos

**Installation:**
```bash
claude mcp add --scope user github -- npx -y @modelcontextprotocol/server-github
```

**Configuration:**
You'll need a GitHub Personal Access Token:
1. Go to https://github.com/settings/tokens
2. Generate new token (classic)
3. Scopes needed: `repo`, `workflow`, `read:org`
4. Save token securely

**Usage Examples:**
- "Show me all open PRs for this repo"
- "What's the status of PR #57?"
- "List all issues labeled 'bug'"
- "Check the CI status for db-guardian-service"

---

### 2. Sequential Thinking MCP (Recommended)
**Why:** Complex architectural decisions across 28 services
**Capabilities:**
- Step-by-step problem decomposition
- Multi-step reasoning for complex tasks
- Trade-off analysis
- Architectural decision making

**Installation:**
```bash
claude mcp add --scope user sequential-thinking -- npx -y @modelcontextprotocol/server-sequential-thinking
```

**Usage Examples:**
- "Help me design a new inter-service communication pattern"
- "What's the best way to implement rate limiting across all services?"
- "Analyze the trade-offs of switching from NATS to Kafka"

---

### 3. Context7 MCP (Recommended)
**Why:** Search documentation for Go, Rust, Node.js, Kubernetes
**Capabilities:**
- Search official documentation
- Find best practices
- Get API references
- Learn framework patterns

**Installation:**
```bash
claude mcp add --scope user context7 -- npx -y context7-mcp
```

**Usage Examples:**
- "What's the latest Fastify authentication pattern?"
- "Show me Go best practices for error handling"
- "How do I configure Kubernetes HPA?"

---

### 4. Postgres MCP (Optional but Useful)
**Why:** Inspect database schemas across services
**Capabilities:**
- Query database schemas
- Inspect table structures
- View indexes and constraints
- Analyze query performance

**Installation:**
```bash
claude mcp add --scope user postgres -- npx -y @modelcontextprotocol/server-postgres
```

**Configuration:**
Connect to your local dev database:
- Host: localhost
- Port: 55432
- Database: lma_dev
- User: lma
- Password: lma123

**Usage Examples:**
- "Show me the schema for the projects table"
- "What indexes exist on the users table?"
- "Find all tables with a foreign key to projects"

---

### 5. Kubernetes MCP (Optional)
**Why:** Inspect Helm charts and K8s resources
**Capabilities:**
- Inspect Kubernetes resources
- Validate YAML manifests
- Check Helm chart values
- Review resource quotas

**Installation:**
```bash
# Requires kubectl configured
claude mcp add --scope user kubernetes -- npx -y @modelcontextprotocol/server-kubernetes
```

**Usage Examples:**
- "Check the resource limits in db-guardian Helm chart"
- "What's the current pod status for projects-service?"

---

## Project-Specific MCP Configuration

For team-wide MCP configuration, create a `.mcp.json` file in the project root:

**File:** `.mcp.json` (commit to git)

```json
{
  "mcpServers": {
    "github": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_TOKEN}"
      },
      "description": "GitHub integration for LMA repository"
    },
    "sequential-thinking": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"],
      "description": "Enhanced reasoning for complex decisions"
    },
    "context7": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "context7-mcp"],
      "description": "Documentation search for Go, Rust, Node, K8s"
    },
    "postgres-dev": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"],
      "env": {
        "POSTGRES_CONNECTION_STRING": "postgresql://lma:lma123@localhost:55432/lma_dev"
      },
      "description": "Local development database (read-only recommended)"
    }
  },
  "defaults": {
    "github": {
      "owner": "SubhajL",
      "repo": "lastmile-accelerator"
    }
  }
}
```

**Environment Variables:**
Create `.env.mcp` (gitignored) for sensitive values:
```bash
GITHUB_TOKEN=ghp_your_token_here
POSTGRES_CONNECTION_STRING=postgresql://lma:lma123@localhost:55432/lma_dev
```

---

## Verifying MCP Installation

### Check Installed Servers
```bash
claude mcp list
```

Expected output:
```
Installed MCP Servers:
  ✓ github
  ✓ sequential-thinking
  ✓ context7
  ✓ postgres-dev
```

### Test Server Connection
```bash
# Test GitHub MCP
claude mcp test github

# Test Postgres MCP
claude mcp test postgres-dev
```

---

## Using MCP Servers with Claude Code

Once installed, MCP servers are automatically available in Claude Code sessions.

### Example: GitHub Integration
```
You: "Show me all open PRs"
Claude: [Uses GitHub MCP to fetch PRs]

PR #57: feat(db): SD-03 add outbox schema DDL and types (inert)
PR #56: test(channels): SD-02 reliability tests + breaker config
PR #55: feat(channels): SD-02 wire per-send reliability
...
```

### Example: Sequential Thinking
```
You: "Help me design a new authentication service"
Claude: [Uses Sequential Thinking MCP]

Let me break this down step-by-step:

Step 1: Requirements Analysis
- Current auth flow uses Keycloak (OIDC)
- Need to support JWT validation across 28 services
- Must integrate with existing Vault secrets

Step 2: Design Considerations
...
```

### Example: Documentation Search
```
You: "What's the latest Fastify v4 authentication pattern?"
Claude: [Uses Context7 MCP to search Fastify docs]

According to Fastify v4 documentation, the recommended authentication pattern is...
```

---

## Advanced Configuration

### Custom MCP Server for LMA

You can create a custom MCP server specific to your project:

**Example: Service Catalog MCP**
```typescript
// tools/mcp-server/service-catalog.ts
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import yaml from 'yaml';
import fs from 'fs';

const server = new Server({
  name: 'lma-service-catalog',
  version: '1.0.0',
}, {
  capabilities: {
    tools: {},
  },
});

server.setRequestHandler('tools/list', async () => ({
  tools: [{
    name: 'get_service_info',
    description: 'Get service details from service_catalog.yaml',
    inputSchema: {
      type: 'object',
      properties: {
        service_name: { type: 'string' }
      },
      required: ['service_name']
    }
  }]
}));

server.setRequestHandler('tools/call', async (request) => {
  if (request.params.name === 'get_service_info') {
    const catalog = yaml.parse(fs.readFileSync('service_catalog.yaml', 'utf8'));
    const service = catalog.services.find(s => s.name === request.params.arguments.service_name);
    return { content: [{ type: 'text', text: JSON.stringify(service, null, 2) }] };
  }
});

server.connect();
```

**Add to `.mcp.json`:**
```json
{
  "mcpServers": {
    "lma-catalog": {
      "type": "stdio",
      "command": "node",
      "args": ["tools/mcp-server/service-catalog.js"],
      "description": "LMA service catalog access"
    }
  }
}
```

---

## Troubleshooting

### MCP Server Not Loading
```bash
# Check MCP server logs
claude mcp logs github

# Restart MCP server
claude mcp restart github
```

### Permission Issues
```bash
# Check MCP server permissions
claude mcp status github

# Re-authenticate
claude mcp remove github
claude mcp add --scope user github -- npx -y @modelcontextprotocol/server-github
```

### Environment Variables Not Working
Ensure `.env.mcp` is in the project root and contains:
```bash
GITHUB_TOKEN=ghp_xxx
```

---

## Security Best Practices

### 1. Token Management
- **DO:** Store tokens in `.env.mcp` (gitignored)
- **DON'T:** Commit tokens to git
- **DO:** Use environment variables in `.mcp.json`
- **DON'T:** Hardcode tokens in `.mcp.json`

### 2. Database Access
- **DO:** Use read-only database user for MCP
- **DON'T:** Use admin credentials
- **DO:** Connect only to development database
- **DON'T:** Connect to production database via MCP

### 3. GitHub Permissions
- **DO:** Use fine-grained tokens with minimal scopes
- **DON'T:** Use tokens with `admin:org` or `delete_repo`
- **DO:** Rotate tokens regularly
- **DON'T:** Share tokens across team members

---

## MCP Server Recommendations by Use Case

### For Daily Development
**Install:** GitHub, Sequential Thinking
```bash
claude mcp add --scope user github -- npx -y @modelcontextprotocol/server-github
claude mcp add --scope user sequential-thinking -- npx -y @modelcontextprotocol/server-sequential-thinking
```

### For Learning Codebase
**Install:** Context7, Postgres
```bash
claude mcp add --scope user context7 -- npx -y context7-mcp
claude mcp add --scope user postgres -- npx -y @modelcontextprotocol/server-postgres
```

### For DevOps Work
**Install:** Kubernetes, GitHub
```bash
claude mcp add --scope user kubernetes -- npx -y @modelcontextprotocol/server-kubernetes
claude mcp add --scope user github -- npx -y @modelcontextprotocol/server-github
```

---

## Next Steps

1. **Install essential MCP servers:**
   ```bash
   claude mcp add --scope user github -- npx -y @modelcontextprotocol/server-github
   claude mcp add --scope user sequential-thinking -- npx -y @modelcontextprotocol/server-sequential-thinking
   ```

2. **Set up GitHub token:**
   - Create token at https://github.com/settings/tokens
   - Add to environment variables

3. **Test installation:**
   ```bash
   claude mcp test github
   ```

4. **Start using in Claude Code:**
   - "Show me all open PRs"
   - "Help me understand the db-guardian architecture"
   - "What's the latest Fastify v4 pattern?"

---

## Resources

- **MCP Documentation:** https://modelcontextprotocol.io
- **Available Servers:** https://github.com/modelcontextprotocol/servers
- **Claude Code MCP Guide:** https://docs.anthropic.com/claude-code/mcp
- **Creating Custom MCP Servers:** https://modelcontextprotocol.io/docs/building-servers

---

**Questions or Issues?**
- File an issue: https://github.com/SubhajL/lastmile-accelerator/issues
- MCP Support: https://github.com/modelcontextprotocol/servers/issues
