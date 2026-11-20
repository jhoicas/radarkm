# RadarKm

Monorepo del copiloto para conductores de plataformas en Colombia.

| Componente | Ruta |
|---|---|
| App Android | `apps/android` |
| Admin Web | `apps/admin-web` |
| API Go | `services/api` |
| Specs SDD | `specs/` |
| Reglas agente | `.agents/rules/` |

Ver `implementation_plan.md` y `AGENTS.md` antes de contribuir.

## Verificación rápida

```bash
cd apps/android && ./gradlew :core:finance:test
cd services/api && go test ./... && go build -o bin/server ./cmd/server
cd apps/admin-web && pnpm install && pnpm exec tsc --noEmit
```
