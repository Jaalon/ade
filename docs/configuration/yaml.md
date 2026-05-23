# Configuration YAML

## Emplacement
Le fichier de configuration peut être placé à :
- `./ade-config.yaml` (racine du projet)
- `./.ade.yaml` (racine du projet)
- `%USERPROFILE%\.ade\config.yaml` (global utilisateur)

## Structure

```yaml
# Configuration des outils agentic
tools:
  opencode:
    path: "C:\\chemin\\personnalise\\opencode.exe"
  cursor:
    path: "C:\\chemin\\personnalise\\Cursor.exe"

# Configuration des serveurs MCP
mcp_servers:
  - name: "filesystem"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "."]
```

## Sections

| Section | Description |
|---------|-------------|
| `tools` | Surcharge des chemins des outils agentic |
| `mcp_servers` | Définition des serveurs MCP à configurer |
