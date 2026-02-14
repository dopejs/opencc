# opencc

[English](../README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md)

Conmutador de entornos multi-CLI para Claude Code, Codex y OpenCode con conmutación automática por fallos en el proxy de API.

## Características

- **Soporte multi-CLI** — Compatible con Claude Code, Codex y OpenCode, configurable por proyecto
- **Gestión multi-configuración** — Gestiona todas las configuraciones de API en `~/.opencc/opencc.json`
- **Conmutación por fallos del proxy** — Proxy HTTP integrado que cambia automáticamente a proveedores de respaldo cuando el principal no está disponible
- **Enrutamiento por escenarios** — Enrutamiento inteligente basado en características de la solicitud (thinking, image, longContext, etc.)
- **Vinculación de proyectos** — Vincula directorios a perfiles y CLIs específicos para configuración automática por proyecto
- **Variables de entorno** — Configura variables de entorno específicas por CLI a nivel de proveedor
- **Interfaz TUI** — Interfaz de terminal interactiva con modos Dashboard y legado
- **Interfaz web de gestión** — Gestión visual desde el navegador para proveedores, perfiles y vinculaciones de proyectos
- **Autoactualización** — Actualización con un solo comando vía `opencc upgrade` con coincidencia de versiones semver
- **Autocompletado de Shell** — Compatible con zsh / bash / fish

## Instalación

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh
```

Desinstalar:

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh -s -- --uninstall
```

## Inicio rápido

```sh
# Abrir la interfaz TUI y crear el primer proveedor
opencc config

# Iniciar (usando el perfil predeterminado)
opencc

# Usar un perfil específico
opencc -p work

# Usar un CLI específico
opencc --cli codex
```

## Referencia de comandos

| Comando | Descripción |
|---------|-------------|
| `opencc` | Iniciar CLI (usando vinculación de proyecto o configuración predeterminada) |
| `opencc -p <profile>` | Iniciar con un perfil específico |
| `opencc -p` | Seleccionar perfil interactivamente |
| `opencc --cli <cli>` | Usar un CLI específico (claude/codex/opencode) |
| `opencc use <provider>` | Usar directamente un proveedor específico (sin proxy) |
| `opencc pick` | Seleccionar interactivamente un proveedor para iniciar |
| `opencc list` | Listar todos los proveedores y perfiles |
| `opencc config` | Abrir la interfaz TUI de configuración |
| `opencc config --legacy` | Usar la interfaz TUI legada |
| `opencc bind <profile>` | Vincular el directorio actual a un perfil |
| `opencc bind --cli <cli>` | Vincular el directorio actual a un CLI específico |
| `opencc unbind` | Eliminar la vinculación del directorio actual |
| `opencc status` | Mostrar el estado de vinculación del directorio actual |
| `opencc web start` | Iniciar la interfaz web de gestión |
| `opencc web open` | Abrir la interfaz web en el navegador |
| `opencc web stop` | Detener el servidor web |
| `opencc upgrade` | Actualizar a la última versión |
| `opencc version` | Mostrar versión |

## Soporte multi-CLI

opencc es compatible con tres CLIs de asistentes de programación con IA:

| CLI | Descripción | Formato de API |
|-----|-------------|----------------|
| `claude` | Claude Code (predeterminado) | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

### Establecer CLI predeterminado

```sh
# Vía TUI
opencc config  # Settings → Default CLI

# Vía Web UI
opencc web open  # Página de Settings
```

### CLI por proyecto

```sh
cd ~/work/project
opencc bind --cli codex  # Usar Codex para este directorio
```

### Usar otro CLI temporalmente

```sh
opencc --cli opencode  # Usar OpenCode para esta sesión
```

## Gestión de perfiles

Un perfil es una lista ordenada de proveedores utilizada para conmutación por fallos.

### Ejemplo de configuración

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-main", "anthropic-backup"]
    },
    "work": {
      "providers": ["company-api"],
      "routing": {
        "think": {"providers": [{"name": "thinking-api"}]}
      }
    }
  }
}
```

### Uso de perfiles

```sh
# Usar perfil predeterminado
opencc

# Usar un perfil específico
opencc -p work

# Selección interactiva
opencc -p
```

## Vinculación de proyectos

Vincula directorios a perfiles y/o CLIs específicos para configuración automática por proyecto.

```sh
cd ~/work/company-project

# Vincular perfil
opencc bind work-profile

# Vincular CLI
opencc bind --cli codex

# Vincular ambos
opencc bind work-profile --cli codex

# Ver estado
opencc status

# Eliminar vinculación
opencc unbind
```

**Prioridad**: Argumentos de línea de comandos > Vinculación de proyecto > Predeterminado global

## Interfaz TUI de configuración

```sh
opencc config
```

v1.5 introduce una nueva interfaz Dashboard:

- **Panel izquierdo**: Proveedores, Perfiles, Vinculaciones de proyectos
- **Panel derecho**: Detalles del elemento seleccionado
- **Atajos de teclado**:
  - `a` - Añadir nuevo elemento
  - `e` - Editar elemento seleccionado
  - `d` - Eliminar elemento seleccionado
  - `Tab` - Cambiar foco
  - `q` - Volver / Salir

Usa `--legacy` para cambiar a la interfaz legada.

## Interfaz web de gestión

```sh
# Iniciar (se ejecuta en segundo plano, puerto 19840)
opencc web start

# Abrir en el navegador
opencc web open

# Detener
opencc web stop
```

Funcionalidades de la interfaz web:
- Gestión de proveedores y perfiles
- Gestión de vinculaciones de proyectos
- Configuración global (CLI predeterminado, perfil predeterminado, puerto)
- Visor de registros de solicitudes
- Autocompletado del campo de modelo

## Variables de entorno

Cada proveedor puede tener variables de entorno específicas por CLI:

```json
{
  "providers": {
    "my-provider": {
      "base_url": "https://api.example.com",
      "auth_token": "sk-xxx",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
        "MAX_THINKING_TOKENS": "50000"
      },
      "codex_env_vars": {
        "CODEX_SOME_VAR": "value"
      },
      "opencode_env_vars": {
        "OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX": "64000"
      }
    }
  }
}
```

### Variables de entorno comunes de Claude Code

| Variable | Descripción |
|----------|-------------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | Tokens de salida máximos |
| `MAX_THINKING_TOKENS` | Presupuesto de pensamiento extendido |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | Ventana de contexto máxima |
| `BASH_DEFAULT_TIMEOUT_MS` | Tiempo de espera predeterminado de Bash |

## Enrutamiento por escenarios

Enruta automáticamente las solicitudes a diferentes proveedores según las características de la solicitud:

| Escenario | Condición de activación |
|-----------|------------------------|
| `think` | Modo thinking activado |
| `image` | Contiene contenido de imagen |
| `longContext` | El contenido supera el umbral |
| `webSearch` | Usa la herramienta web_search |
| `background` | Usa el modelo Haiku |

**Mecanismo de fallback**: Si todos los proveedores en la configuración del escenario fallan, se recurre automáticamente a los proveedores predeterminados del perfil.

Ejemplo de configuración:

```json
{
  "profiles": {
    "smart": {
      "providers": ["main-api"],
      "long_context_threshold": 60000,
      "routing": {
        "think": {
          "providers": [{"name": "thinking-api", "model": "claude-opus-4-5"}]
        },
        "longContext": {
          "providers": [{"name": "long-context-api"}]
        }
      }
    }
  }
}
```

## Archivos de configuración

| Archivo | Descripción |
|---------|-------------|
| `~/.opencc/opencc.json` | Archivo de configuración principal |
| `~/.opencc/proxy.log` | Registro del proxy |
| `~/.opencc/web.log` | Registro del servidor web |

### Ejemplo de configuración completa

```json
{
  "version": 5,
  "default_profile": "default",
  "default_cli": "claude",
  "web_port": 19840,
  "providers": {
    "anthropic": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000"
      }
    }
  },
  "profiles": {
    "default": {
      "providers": ["anthropic"]
    }
  },
  "project_bindings": {
    "/path/to/project": {
      "profile": "work",
      "cli": "codex"
    }
  }
}
```

## Actualización

```sh
# Última versión
opencc upgrade

# Versión específica
opencc upgrade 1.5
opencc upgrade 1.5.0
```

## Migración desde versiones anteriores

Si usabas anteriormente el formato `~/.cc_envs/`, opencc migrará automáticamente a `~/.opencc/opencc.json`.

## Desarrollo

```sh
# Compilar
go build -o opencc .

# Probar
go test ./...
```

Publicación: Empuja un tag y GitHub Actions compilará automáticamente.

```sh
git tag v1.5.1
git push origin v1.5.1
```

## License

MIT
