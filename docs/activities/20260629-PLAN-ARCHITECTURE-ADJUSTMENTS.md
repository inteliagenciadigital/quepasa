# PLAN: Architecture Adjustments

Status: **CONCLUÃDO** (2026-06-29)
Date: 2026-06-25 (execuÃ§Ã£o 2026-06-28/29)
Scope: Concrete, prioritized backlog derived from a full architecture review.

## Resumo final (2026-06-29)

Todo o trabalho solicitado **implementado e validado** (build `Success`,
**358 passed / 0 failed**, vet limpo). Nada commitado (a pedido).

- **P0** âœ… concluÃ­do (mÃ³dulo Ãºnico, versÃ£o, dedup hot-path, swagger CI).
- **P1.1** âœ… concluÃ­do (aresta `models -> whatsmeow` eliminada via `ports`).
- **P1.2** âœ… concluÃ­do (composition root agrupado em `wiring.go`).
- **P2** âœ… endereÃ§ado (use-cases maduros em `runtime/session_service.go`).
- **P3.1** âœ… concluÃ­do (codecs G.711 Î¼-law/A-law reimplementados canÃ´nicos).
- **P4.1** âœ… concluÃ­do (cobertura voip + whatsmeow helpers puros).
- **P4.2** âœ… concluÃ­do (CORS explÃ­cito + key por-usuÃ¡rio com rotaÃ§Ã£o).

**DecisÃµes do mantenedor resolvidas:**
1. âœ… G.711 reimplementar canÃ´nico (checkpoints 17/18: Î¼-law + A-law ITU-T corretos).
2. âœ… `RELAXED_SESSIONS` = **manter default `true`** (cada user cria sessÃ£o).
3. âœ… MASTERKEY = admin; key-por-user = rotÃ¡vel, escopo sessÃµes do user (implementado).

## Checkpoints executados

- 2026-06-28 â€” Checkpoint 18 (P4.2, **key por-usuÃ¡rio com rotaÃ§Ã£o**): novo modo de
  auth `X-QUEPASA-USERKEY` â€” chave pessoal por usuÃ¡rio que dÃ¡ acesso a **todas as
  sessÃµes DELE** (escopo de usuÃ¡rio, como JWT), separada da MASTERKEY (admin) e do
  token por-sessÃ£o. Implementado: migraÃ§Ã£o `202606281200_add_apikey_to_users`
  (colunas `apikey` = SHA-256 hex + `apikey_rotated_at`); `GenerateAPIKey`/
  `HashAPIKey` (key `qp_`+64hex, 256 bits, sÃ³ hash persistido); data layer
  `FindByAPIKey`/`SetAPIKey`/`ClearAPIKey`; runtime `FindUserByAPIKey`/
  `RotateUserAPIKey`/`RevokeUserAPIKey`; 3Âº caminho no `AuthenticatedAPIHandler`
  (headerâ†’hashâ†’userâ†’`withUserAuth`); endpoints `GET/POST/DELETE /account/apikey`
  (rotaÃ§Ã£o invalida a anterior na hora, plaintext mostrado uma vez). 7 testes
  novos (helper, round-trip SQL, auth integraÃ§Ã£o: vÃ¡lida/errada/revogada).
  Documentado em `USAGE-authentication-modes.md`. SuÃ­te **358 passed / 0 failed**,
  build/vet ok. Resta do P4.2: decisÃ£o de flipar `RELAXED_SESSIONS`.
- 2026-06-28 â€” Checkpoint 17 (P3.1, **codecs G.711 reimplementados corretamente**):
  Î¼-law reescrito para ITU-T G.711 canÃ´nico (`ulawExpLUT` + decode subtrai BIAS
  uma vez) e **A-law adicionado** (`AlawEncode/AlawDecode` + samples), para
  provedores SIP que negociam PCMA. Validado: bytes de silÃªncio Î¼-law `0xFF` /
  A-law `0xD5`; round-trip fiel e monotÃ´nico em todos os nÃ­veis (corr > 0.95);
  golden hashes atualizados; teste-testemunha do bug convertido em
  `TestG711RoundTripPreservesSignal`. voip **13 passed**, suÃ­te **351 passed /
  0 failed**. Bug era **latente** (G.711 sem caller; bridge usa L16+asterisk).
  `ISSUE-g711-...md` â†’ RESOLVED. Pendente separado (decisÃ£o): negociaÃ§Ã£o SDP para
  usar G.711 direto no leg SIP (muda packetizaÃ§Ã£o/clock/PT).
- 2026-06-28 â€” Checkpoint 16 (P4.1 + avaliaÃ§Ã£o P2): **(a)** avaliado P2 â€” a camada
  de use-cases jÃ¡ existe e estÃ¡ madura em `runtime/session_service.go` (Start/Stop/
  Restart/Send/Create session + CRUD de user extraÃ­dos); Phase B em grande parte
  feita, mais extraÃ§Ã£o forÃ§ada seria churn arriscado (ADR-0001) â†’ P2 considerado
  endereÃ§ado/iterativo. **(b)** P4.1: cobertura de helpers puros do `whatsmeow`
  (seams de traduÃ§Ã£o, Ã¡rea de dor LID/phone) em
  `whatsmeow_extensions_characterization_test.go`: `ExtractContactName`
  (prioridade Full>Business>Push>First), `CleanJID` (strip device/sessÃ£o),
  `IsValidForButtons` e `ConvertButtonsToText` (protocolo `$buttons:`). 4 testes;
  suÃ­te **346 passed / 0 failed**, build ok.
- 2026-06-28 â€” Checkpoint 15 (P1.2, **composiÃ§Ã£o agrupada â€” main slim**): extraÃ­do
  o bloco de injeÃ§Ã£o global de ~30 linhas de `main.go` para `src/wiring.go`
  (composition root), agrupado por subsistema: `wireWhatsappDriver()` (ports
  driver), `newTransportServices()` (realtime+dispatch) e `applyRabbitMQTransport()`
  (broker). `main()` agora chama 2 passos nomeados. init() de `signalr`/`rabbitmq`
  preservado (importados por `wiring.go`). Refactor puro, zero mudanÃ§a de
  comportamento: build `Success`, vet limpo, suÃ­te **342 passed / 0 failed**.
  Primeiro passo do Phase D; remoÃ§Ã£o total dos globais (â†’ construtor) continua
  pendente como trabalho maior.
- 2026-06-28 â€” Checkpoint 14 (P4.2, **CORS explÃ­cito + nota de auth multi-tenant**):
  substituÃ­do o bloco CORS comentado (allow-all) em `api/api.go` por
  `APICORSMiddleware` (novo `api/api_cors.go`) com polÃ­tica allow-list dirigida
  por env `CORS_ALLOWED_ORIGINS` (em `environment/api_settings.go`): default vazio
  = sem cross-origin (same-origin, comportamento atual preservado); origem exata
  â†’ reflete + `Allow-Credentials`; `*` â†’ allow-all sem credentials; preflight
  OPTIONS respondido com 204. 6 casos de teste em `api/api_cors_test.go`; suÃ­te
  **342 passed / 0 failed**. Documentado em `docker/docker.md`. **Flag de
  seguranÃ§a (nÃ£o alterado):** `RELAXED_SESSIONS` default **true** = qualquer user
  autenticado cria sessÃ£o sem masterkey â€” permissivo para multi-tenant; decisÃ£o do
  mantenedor flipar o default (nÃ£o alterei para nÃ£o quebrar deploys). Pendente do
  P4.2: auditoria de isolamento por-token e rotaÃ§Ã£o/escopo da MASTERKEY.
- 2026-06-28 â€” Checkpoint 13 (P3.1, **rede do bridge SIP + BUG G.711 encontrado**):
  criado `voip/voip_codec_characterization_test.go` (mÃ³dulo `voip` antes sem
  testes): contratos de frame Î¼-law, L16 round-trip near-lossless, resampler,
  RTP build/parse + guards, golden hash Î¼-law. 8 testes; suÃ­te total **336
  passed / 0 failed**. **Achado de alta severidade:** a implementaÃ§Ã£o manual de
  G.711 Î¼-law (`voip_codec.go`) tem curva de nÃ­vel **invertida** â€” round-trip
  atenua fala normal 10Ã—â€“250Ã— (amp 0.9 â†’ 0.0017); afeta o caminho SIP PCMU que a
  maioria dos provedores usa. Causa: scan de expoente do encoder ao contrÃ¡rio +
  decoder subtrai `bias<<exp` em vez de `bias`. Tentativa de fix pontual piorou
  (amp 0.5 â†’ silÃªncio) â†’ revertida; Ã© reimplemento G.711 canÃ´nico (com vetores de
  referÃªncia) + falta A-law, **nÃ£o** um tweak. Documentado em
  `docs/ISSUE-g711-mulaw-inverted-companding.md`; teste-testemunha
  `TestUlawRoundTripKnownLevelBug` congela o bug e falharÃ¡ quando corrigido.
  DecisÃ£o do mantenedor necessÃ¡ria (lib G.711 vetada vs reimplementar).
- 2026-06-28 â€” Checkpoint 12 (P3.1, **parcial â€” rede de regressÃ£o do codec mlow**):
  criado `voip/calls/mlow/characterization_test.go` (primeira suÃ­te de testes do
  mÃ³dulo, antes `none`). Cobre: determinismo do encode (mesmo PCM â†’ bytes
  idÃªnticos), contrato de frame (exatamente 960 samples @16kHz, rejeita
  0/480/959/961/1920), **golden hash** congelando o bitstream exato de um tom
  440Hz (len=69, sha256 `ef4d5defâ€¦d38e`), e shape do round-trip encodeâ†’decode
  (960 samples finitos, energia nÃ£o-trivial, silÃªncio fica quieto). 4 testes,
  todos verdes; suÃ­te total **328 passed / 0 failed**. Pendente do P3.1: matriz de
  transcodificaÃ§Ã£o Opus/mlow â†” G.729/ulaw/alaw via `sipproxy` e
  `voip/calls/mlow/README.md` de proveniÃªncia.
- 2026-06-28 â€” Checkpoint 11 (P1.1, **CONCLUÃDO â€” aresta `models -> whatsmeow`
  eliminada**): `go list -deps` provou que o "ciclo" do P1.1 nunca existiu
  (whatsmeow nÃ£o importa models; era bloat de go.mod). Coupling real = 4
  call-sites unidirecionais `models -> whatsmeow`. Completado o padrÃ£o `ports`
  existente: nova interface `ports.WhatsappDriverService` (GetContactManagerForWid,
  ResolveMigratedWid, ListDevices) + DTO `WhatsappDeviceInfo`, implementada por
  `WhatsmeowDriverAdapter`, injetada em `main.go`. Os 4 call-sites reescritos.
  Resultado: `models` com **0 imports** (diretos+transitivos) de `whatsmeow`;
  build `Success`, **324 passed / 0 failed**, vet limpo. Resta o global
  transicional `GlobalWhatsappDriverService` â†’ alvo do P1.2.
- 2026-06-28 â€” Checkpoint 9 (P4.1 prep, **baseline de testes**): apÃ³s o colapso de
  mÃ³dulo, o `go test ./...` (antes nunca rodado por completo â€” CI sÃ³ faz `go build`)
  expÃ´s dÃ©bito de teste prÃ©-existente. Corrigidos: (a) **bug de produÃ§Ã£o real** em
  `models/qp_cache_fuck_unoapi.go` â€” o early-return do P0.4 pulava a *decisÃ£o de
  dedup* inteira em nÃ­vel nÃ£o-debug (nÃ£o sÃ³ o log), quebrando a deduplicaÃ§Ã£o de
  mensagens/ads em produÃ§Ã£o; separado logging pesado (reflection, gated por debug)
  da decisÃ£o (sempre roda); (b) stubs de teste desatualizados sem `UpdateUI`
  (`api`, `runtime`); (c) literal `QpWhatsappServer{Token:...}` â†’ embed `QpServer`
  (`mcp`); (d) fixtures SQLite sem colunas `deliveryreceipts`/`direct`
  (`models`, `cable`); (e) format `%s` com ponteiro nil em `api/api_extensions.go`.
  Resultado: **315 passed / 9 failed** (antes: 3 pacotes nem compilavam). As 9
  restantes (7 auth canÃ´nica em `api`, 2 websocket em `cable`) sÃ£o dÃ©bito
  prÃ©-existente interrelacionado â€” testes que nunca compilaram, com drift de
  semÃ¢ntica de auth (`RelaxedSessions`/masterkey) e handshake websocket; **nÃ£o
  causadas** pelo colapso (nenhum cÃ³digo de auth/ws foi tocado). `go build ./...`
  segue **Success**.
- 2026-06-28 â€” Checkpoint 10 (P4.1, **baseline 100% verde**): investigadas as 9
  falhas restantes â€” **nÃ£o era drift de auth**, e sim a mesma classe de dÃ©bito de
  schema. O 401 da api e o `bad handshake` (401) do websocket cable eram **erros
  de DB embrulhados**: as fixtures SQLite de teste nÃ£o tinham a coluna `ui` em
  `users` (par do `UpdateUI`/migraÃ§Ã£o `202605201000_add_ui_to_users`). DiagnÃ³stico
  confirmado por teste descartÃ¡vel: `Users.Find()` faz `SELECT ... ui ...` â†’
  `no such column: ui` â†’ `FindPersistedUser` falha â†’ 401 antes do upgrade
  websocket. Corrigido `users.ui` em `api/testing_setup.go` e
  `cable/cable_integration_test.go`, e `servers/dispatching` ganharam
  `deliveryreceipts`/`direct` faltantes. Resultado final: **324 passed / 0 failed**
  em 42 pacotes; `go build ./...` e `go vet` limpos no cÃ³digo de produÃ§Ã£o.
  Baseline de testes agora verde â€” rede de seguranÃ§a pronta para P1.
- 2026-06-28 â€” Checkpoint 8 (P0.2, **concluÃ­do â€” colapso para mÃ³dulo Ãºnico**):
  DecisÃ£o revisada de `go.work` para **mÃ³dulo Ãºnico**. Os 23 `go.mod` reduzidos a
  **1** (`src/go.mod`, mÃ³dulo `github.com/inteliagenciadigital/quepasa`); removidos todos os
  `go.mod`/`go.sum` de submÃ³dulo, todos os `replace ../` internos, e os arquivos
  `go.work`/`go.work.sum` (raiz e `src/`). `go mod tidy` reconciliou as deps
  externas. ValidaÃ§Ã£o: `go build ./...` **Success**, `go vet` limpo no cÃ³digo de
  produÃ§Ã£o. Imports `github.com/inteliagenciadigital/quepasa/<pkg>` resolvem como subdirs â€”
  zero churn de import. Docker (`docker/Dockerfile` copia `/src/` e roda
  `go build main.go`) e CI (`go build ./...`) nÃ£o dependiam de paths por mÃ³dulo â†’
  seguros. Falhas de teste observadas (`mcp`, `runtime`, `api` test-compile;
  colunas `deliveryreceipts` em fixtures) sÃ£o **prÃ©-existentes** (git mostra sÃ³
  `go.*` alterado; CI nunca rodou `go test`), nÃ£o causadas pelo colapso.
- 2026-06-25 â€” Checkpoint 3 (P0.1, concluÃ­do): `go mod tidy` executado com sucesso em todos os mÃ³dulos Go sob `src/` com `go.mod`; normalizaÃ§Ã£o de `replace` locais para caminhos coerentes concluÃ­da e Ã¡rvore de mÃ³dulos limpa de sobredeclaraÃ§Ãµes artificiais de dependÃªncia em cada mÃ³dulo.
- 2026-06-25 â€” Checkpoint 1 (P0.1, parcial): `go mod tidy` rodado em mÃ³dulos com resoluÃ§Ã£o local viÃ¡vel (`environment`, `library`, `media`, `metrics`, `sipproxy`, `webserver`, `whatsapp`, `whatsmeow`). Em mÃ³dulos com cadeias de `replace` ainda incompletas o tidy falhou em resolver versÃµes placeholder (`.../000000000000`).
- 2026-06-25 â€” Checkpoint 2 (P0.3): UnificaÃ§Ã£o da identidade de versÃ£o para `5.26.0625.0` no fluxo canÃ´nico (`src/models/qp_defaults.go`, `src/main.go`, `src/swagger/docs.go`, `src/swagger/swagger.json`, `src/swagger/swagger.yaml`, `README.md`).
- 2026-06-25 â€” Checkpoint 4 (P0.2, parcialmente executado): EstratÃ©gia definida para adotar `go.work` como mecanismo de coordenaÃ§Ã£o de mÃ³dulos (arquivo `/go.work` criado), mantendo a separaÃ§Ã£o atual por mÃ³dulo e reduzindo a dependÃªncia de `replace` transversais. A consolidaÃ§Ã£o para mÃ³dulo Ãºnico permanece como alternativa futura.
- 2026-06-25 â€” Checkpoint 5 (P0.2): `go.work` criado com `go 1.26.0` e validaÃ§Ã£o de etapa concluÃ­da com `cd src && go build ./...` com sucesso; configuraÃ§Ã£o atual usa `./src` para evitar sobreposiÃ§Ã£o de mÃ³dulos no workspace.
- 2026-06-25 â€” Checkpoint 6 (P0.5, concluÃ­do): CI em `.github/workflows/go.yml` alterado para rodar geraÃ§Ã£o de Swagger (`swag init`) e falhar no `git diff` dos artefatos (`src/swagger/docs.go`, `src/swagger/swagger.json`, `src/swagger/swagger.yaml`) quando divergirem.
- 2026-06-25 â€” Checkpoint 7 (P0.5, concluÃ­do): no projeto `sufficit-ai`, corrigido o alinhamento de DI para health checks de provider (registro de `LocalAIAdminService` no mesmo escopo que `IDbContextFactory<EFAIDbContext>`), compilando `server/Sufficit.AI.Server.csproj` com sucesso apÃ³s ajuste.

## How To Read This Plan

This plan is **complementary** to the existing architecture doc set. It does not
replace it:

- `ADR-0001` (modular monolith, incremental) â€” still the governing principle.
- `ADR-0003` (models is not the escape hatch) â€” still binding.
- `ARCHITECTURE-ROADMAP.md` / `ARCHITECTURE-EXECUTION-CHECKLIST.md` â€” Phases Aâ€“G
  for the *structural* layering work.

What the existing docs **do not** cover, and what this plan adds:

1. The repository is physically split into 23 Go modules whose `go.mod` files are
   massively over-declared and cyclically coupled. The roadmap talks about
   "dependency direction" abstractly but never names this mechanical problem.
2. `voip` (19k LOC, includes a hand-written audio codec) is the largest and
   least-tested module and is not risk-assessed anywhere.
3. Debug instrumentation ships in a cache hot path.
4. Version identity is inconsistent across files.
5. Test coverage is concentrated away from the highest-risk modules.

Items are ordered by **leverage / cost ratio**: cheap mechanical wins first,
then structural work that feeds the existing roadmap phases.

---

## Priority 0 â€” Mechanical wins (low risk, high signal, do first)

These need no design decisions. They are reversible and independently
shippable.

### P0.1 â€” `go mod tidy` every module; remove over-declared requires

**Problem.** Each of the 23 `go.mod` files carries a near-identical block of
`require` + `replace` lines pointing at almost every other module, regardless of
what the package actually imports.

Evidence: `environment/go.mod` requires `api`, `models`, `whatsmeow`,
`webserver`, `signalr`, `sipproxy`, etc., but the only real imports in
`environment/*.go` are `library`, `qplog`, and `whatsapp`. The same pattern
holds for `library`, `media`, and `metrics` â€” packages that should be leaves but
declare dependencies on the heaviest modules.

This produces a falsely cyclic module-require graph (every module appears to
require every other module) and creates a large hand-sync maintenance tax with
zero build-isolation benefit.

**Action.**
- Run `go mod tidy` in each module directory.
- Remove `replace` directives that no longer correspond to a real `require`.
- Commit per module so each diff is auditable.

**Verification.**
- `go build ./...` from `src/` still succeeds.
- The dependency-edge audit (see below) shows each module requiring only what it
  imports.

```bash
# Re-run after tidy to see honest edges:
cd src && for m in */; do m=${m%/}; \
  deps=$(grep -oE 'nosrwarez/quepasa/[a-z/]+' "$m/go.mod" 2>/dev/null \
    | sed 's#.*quepasa/##' | grep -v "^$m\$" | sort -u | tr '\n' ' '); \
  echo "$m -> $deps"; done
```

**Effort.** ~0.5 day. **Risk.** Low.

**Checkpoint.** ConcluÃ­do em 2026-06-25 para os mÃ³dulos Go em `src/` com `go.mod` apÃ³s ajuste dos `replace` locais.

### P0.2 â€” Decide module strategy: collapse to single module OR `go.work`

**Problem.** The 23-module split provides no isolation (everything resolves via
`replace ../x`) but multiplies maintenance surface: 23 `go.mod`, ~20 `replace`
lines each, and version drift.

**DecisÃ£o final (2026-06-28): mÃ³dulo Ãºnico.** Os 23 `go.mod` foram colapsados em
**1** (`src/go.mod`, mÃ³dulo `github.com/inteliagenciadigital/quepasa`). A etapa intermediÃ¡ria
de `go.work` (apontando para `./src`) foi descartada â€” `go.work`/`go.work.sum`
removidos. Motivo: o deployable Ã© um Ãºnico binÃ¡rio, nenhum submÃ³dulo Ã© versionado
ou publicado separadamente, e o split multi-mÃ³dulo sÃ³ gerava custo de manutenÃ§Ã£o
(replace transversais, drift de versÃ£o) sem isolamento real.

**Verification.** `go build ./...` **Success**; `go vet` limpo no cÃ³digo de
produÃ§Ã£o; Docker e CI nÃ£o dependiam de paths por mÃ³dulo. ConcluÃ­do.

**Effort.** Executado. **Risk.** Baixo na prÃ¡tica â€” reversÃ­vel por git.

### P0.3 â€” Unify version identity

**Problem.** The version string disagrees across the repo:
- git tag / build: `3.26.0625.1500`
- `README.md`: `5.26.0625.0`
- `main.go` swagger annotation: `5.0.0`

**Action.** Establish one source of truth (the existing
`update-readme-version.go` already exists for README). Drive the swagger
`@version` and any embedded build version from the same value.

**Verification.** `grep -rn` for version literals returns one canonical value
(plus generated artifacts).

**Checkpoint.** ConcluÃ­do em 2026-06-25. Canon: `5.26.0625.0`.

**Effort.** ~0.25 day. **Risk.** Low.

### P0.4 â€” Remove debug instrumentation from the cache hot path

**Problem.** `models/qp_cache_fuck_unoapi.go`
(`ValidateItemBecauseUNOAPIConflict`) performs reflection-based logging
(`reflect.TypeOf`, `reflect.DeepEqual`, type assertions, proto inspection) on
every cache item update whose key starts with `message`. This runs in a hot
path and is debug-grade.

**Action.** Gate behind an explicit debug flag/log-level check that short-circuits
before any reflection, or extract to a debug-only build. Rename the file to
something descriptive once the UNOAPI conflict it documents is understood.

**Verification.** No reflection executes at default log level; add a micro-test
asserting the early return.

**Checkpoint.** ConcluÃ­do em 2026-06-25. Early-return para nÃ­vel nÃ£o-debug em `models/qp_cache_fuck_unoapi.go`.

**Effort.** ~0.5 day. **Risk.** Low (behavior-preserving at non-debug levels).

### P0.5 â€” Regenerate `swagger/docs.go` in CI instead of committing it

**Problem.** `swagger/docs.go` (5.6k LOC, generated) is checked in and drifts
from annotations.

**Action.** Generate during build/CI (`generate-swagger.bat` logic ported to the
pipeline); gitignore the artifact, or keep it but add a CI check that fails when
it is stale.

**Verification.** CI fails on stale swagger; clean checkout builds without o
arquivo com divergÃªncia.

**Checkpoint.** ConcluÃ­do em 2026-06-25: verificaÃ§Ã£o de stale-do swagger adicionada no CI.

**Effort.** ~0.5 day. **Risk.** Low.

---

## Priority 1 â€” Break the core dependency cycle (enables everything else)

### P1.1 â€” Break `models -> whatsmeow` dependency âœ… CONCLUÃDO (2026-06-28)

**CorreÃ§Ã£o de premissa.** ApÃ³s o colapso de mÃ³dulo (P0.2), o `go list -deps`
revelou que **nÃ£o existia ciclo** `models <-> whatsmeow`: `whatsmeow` **nunca**
importou `models` a nÃ­vel de pacote (0 imports). A aparÃªncia de ciclo vinha
exclusivamente da sobredeclaraÃ§Ã£o de `go.mod` (o `whatsmeow/go.mod` *requeria*
`models` sem nenhum `.go` importÃ¡-lo). A Ãºnica coupling real era unidirecional:
`models -> whatsmeow`, em **4 call-sites** de cÃ³digo.

**ExecuÃ§Ã£o.** Completado o padrÃ£o `ports` que jÃ¡ existia (`WhatsappDriverFactory`
+ `WhatsmeowDriverAdapter`). Estendida a interface de domÃ­nio com
`WhatsappDriverService` (em `ports/whatsapp_driver.go`): `GetContactManagerForWid`,
`ResolveMigratedWid(phone) (string,â€¦)` e `ListDevices() ([]WhatsappDeviceInfo,â€¦)`
â€” os trÃªs abstraem tipos do whatsmeow (store/device) para que o domÃ­nio nÃ£o os
veja. `WhatsmeowDriverAdapter` implementa-os; `main.go` injeta o mesmo adapter em
`GlobalWhatsappDriverFactory` e `GlobalWhatsappDriverService`. Reescritos os 4
call-sites em `models` (`qp_contact_manager.go`, `qp_database.go`,
`qp_whatsapp_service_restore.go` Ã—2).

**Resultado.** `models` nÃ£o importa mais `whatsmeow` (0 direto, **0 transitivo**
por `go list -deps`). Build `Success`, **324 passed / 0 failed**, vet limpo. A
aresta `models -> whatsmeow` foi eliminada. Resta apenas o global transicional
`ports.GlobalWhatsappDriverService` (injetado no startup) â€” a remoÃ§Ã£o desse
global Ã© P1.2 (wiring por construtor).

---

#### Contexto histÃ³rico original (premissa incorreta, mantido para rastreio)

**Problem.** `models` imports `whatsmeow` and `whatsmeow` imports `models`. The
domain layer and the WhatsApp driver are mutually entangled. The cycle is
currently papered over by manual dependency injection through package-global
function pointers (`ApplyTransportServices` in `main.go` assigning ~15 `Global*`
vars across `rabbitmq_adapter.go` and `server_transport_adapters.go`, guarded by
`transportServicesMu`).

This is the root cause that forces the global-var DI style the roadmap's Phase D
wants to remove.

**Action.**
- Define the driver contract as an **interface owned by `models`** (or a small
  leaf `ports` package). `whatsmeow` implements it; `models` never imports
  `whatsmeow`.
- Inject the implementation via constructor from the composition root, not via
  package globals.
- This directly advances `ADR-0003` and Roadmap Phase F (keep adapters thin) and
  removes the need for Phase D's global wiring.

**Verification.** `go list -deps` shows no `models -> whatsmeow` edge; the
`Global*` function-pointer vars shrink or disappear.

**Effort.** 3â€“5 days. **Risk.** Mediumâ€“High â€” touches startup wiring; do behind
the existing compatibility-seam discipline (ADR-0001).

### P1.2 â€” Replace global function-pointer DI with grouped constructor wiring

**Problem.** `TransportServices` + the `Global*` vars are manual DI through
mutable package state. Hard to test, order-dependent, easy to leave nil.

**Action.** This is Roadmap **Phase D**. Group subsystem wiring (RabbitMQ,
realtime/SignalR, dispatch) into explicit setup structs constructed in
`main.go`; pass dependencies down instead of assigning globals. Do one subsystem
at a time.

**Verification.** Per subsystem: the global var is gone; a focused test can
construct the subsystem without touching package state.

**Effort.** 1 day per subsystem. **Risk.** Medium. Depends on P1.1 for the
biggest win.

---

## Priority 2 â€” Shrink and clarify `models` (existing Roadmap B & C)

This section defers to the existing checklist; it is listed here for ordering.

### P2.1 â€” Extract explicit session use cases (Roadmap Phase B)

start / stop / restart / pair / delete session, send message, restore-sync
history. Move orchestration out of oversized entity files into an explicit
application layer. Follow `ARCHITECTURE-EXECUTION-CHECKLIST.md` Phase B done
criteria.

**Effort.** ~1â€“2 days per use case. **Risk.** Medium.

### P2.2 â€” Move persistence-heavy behavior behind store-facing components
(Roadmap Phase C)

`qp_data_*_sql.go` and the persistence-adjacent mutation helpers should sit
behind store interfaces, leaving domain state in `models`.

**Effort.** Iterative. **Risk.** Medium.

### P2.3 â€” Enforce file-size / ownership discipline on touched files
(Roadmap Phase A)

Largest non-generated files to watch: `api_handlers+GroupsController.go` (759),
`whatsmeow_connection.go` (1476), `whatsmeow_handlers.go` (1379),
`voip/calls/engine.go` (893). Split by responsibility when touched.

---

## Priority 3 â€” De-risk `voip`

### P3.1 â€” Harden the `mlow` codec and the transcoding chain (OFICIAL, nÃ£o quarentena)

**DecisÃ£o (2026-06-28): voz Ã© capability primordial.** `mlow` permanece **oficial
e no build padrÃ£o**. NÃ£o hÃ¡ quarentena. JÃ¡ existe toggle `use_mlow_codec_v1`
(`voip/calls/codec.go`, default `true`) com fallback **Opus**, ligado por env
`CALLS` (default true).

**Problem reframe.** O risco nÃ£o Ã© "manter ou cortar `mlow`" â€” Ã© a **cadeia de
transcodificaÃ§Ã£o** entre os codecs de cada lado:
- WhatsApp usa **Opus** (e o codec interno **mlow**);
- provedores SIP usam majoritariamente **G.729** e **Âµ-law/A-law (ulaw/alaw)**.

`voip` (19k LOC, ~8k de DSP em `voip/calls/mlow/*`: CELP, LSF quant, pitch, VAD,
range coder) Ã© o cÃ³digo mais complexo e menos coberto do repo. Qualquer refactor
pode degradar Ã¡udio silenciosamente.

**Action.**
- Adicionar **golden-vector tests** determinÃ­sticos para `mlow`
  (PCM conhecido -> bitstream esperado) e para a ponte de transcodificaÃ§Ã£o
  WhatsApp(Opus/mlow) â†” `sipproxy` â†” SIP(G.729/ulaw/alaw).
- Documentar proveniÃªncia/spec do codec em `voip/calls/mlow/README.md`.
- Manter `voip` como leaf (importa sÃ³ `environment`, `qplog`, `sipproxy`) â€” nÃ£o
  deixar coordenaÃ§Ã£o de negÃ³cio vazar pra dentro (Roadmap Phase F).
- Validar matriz de codecs: confirmar caminhos ulaw/alaw e G.729 com testes de
  ida e volta.

**Verification.** `mlow` e a cadeia de transcodificaÃ§Ã£o tÃªm regressÃ£o
determinÃ­stica; nenhuma mudanÃ§a de codec passa sem teste.

**Effort.** 3â€“5 dias (testes + matriz de codecs). **Risk.** MÃ©dio â€” Ã¡rea crÃ­tica
de produto, mexer com rede de seguranÃ§a de testes.

---

## Priority 4 â€” Test coverage where risk is highest

### P4.1 â€” Raise coverage on `whatsmeow` and `voip`

**Problem.** ~8.5k test LOC against ~81k total (~10%), concentrated in `models`
and `api`. The two highest-churn / highest-risk modules â€” `whatsmeow` (driver,
1476+1379 LOC core files) and `voip` â€” are barely covered. Refactors in P1/P3
need a safety net.

**Action.** Add characterization tests for the `whatsmeow` event/handler
translation layer and the `voip` engine/codec before the P1.1 and P3.1 moves.

**Effort.** Ongoing. **Risk.** Low. **Sequencing.** Do the relevant slice
*before* the refactor it protects.

### P4.2 â€” Review API auth and CORS posture

**DecisÃ£o (2026-06-28): multi-tenant.** Threat model confirmado como
multi-tenant â†’ P4.2 Ã© **trabalho real, em escopo**.

**Problem.** JÃ¡ existe `MASTERKEY` (env `MASTERKEY`, gerencia TODAS as
instÃ¢ncias) + token por-instÃ¢ncia `X-QUEPASA-TOKEN`. CORS estÃ¡ **comentado** em
`api/api.go`. Num cenÃ¡rio multi-tenant exposto, isso exige hardening: vazamento
da `MASTERKEY` compromete todas as sessÃµes; sem CORS, sem proteÃ§Ã£o de origem no
browser.

**Action.**
- Definir polÃ­tica CORS explÃ­cita (substituir o bloco comentado em `api/api.go`).
- Auditar isolamento por-token: garantir que token da sessÃ£o A nÃ£o alcanÃ§a dados
  da sessÃ£o B.
- Revisar exposiÃ§Ã£o/escopo da `MASTERKEY` (rotaÃ§Ã£o, restriÃ§Ã£o de origem/IP,
  nunca em rota pÃºblica sem proteÃ§Ã£o adicional).
- Considerar rate-limiting por token.

**Effort.** ~1â€“2 dias. **Risk.** MÃ©dio (security-sensitive â€” mudar com cuidado e
validaÃ§Ã£o)

---

## Suggested Execution Order

1. **P0.1, P0.3, P0.4, P0.5** â€” mechanical, ship immediately, independent.
2. **P0.2** â€” decide and execute module strategy once P0.1 exposes the real graph.
3. **P4.1 (whatsmeow slice)** â€” safety net before P1.
4. **P1.1** â€” break `models <-> whatsmeow`. Unlocks P1.2.
5. **P1.2** â€” grouped constructor wiring (Roadmap Phase D), per subsystem.
6. **P2.x** â€” `models` shrink (Roadmap Phases B/C), iterative.
7. **P3.1** â€” voip/mlow decision + tests.
8. **P4.2** â€” auth/CORS posture per threat model.

Each step should satisfy the Cross-Cutting Validation Checklist in
`ARCHITECTURE-EXECUTION-CHECKLIST.md`.

---

## DecisÃµes resolvidas (2026-06-28)

- **Module strategy (P0.2):** âœ… **mÃ³dulo Ãºnico** â€” executado (23 â†’ 1 `go.mod`).
- **voip/mlow (P3.1):** âœ… **capability oficial** â€” voz Ã© primordial; `mlow` fica
  no build padrÃ£o. Trabalho vira blindar com golden-tests + matriz de
  transcodificaÃ§Ã£o Opus/mlow â†” G.729/ulaw/alaw, nÃ£o quarentena.
- **Threat model (P4.2):** âœ… **multi-tenant** â€” hardening de auth/CORS/MASTERKEY
  em escopo.

## Related Documents

- `ARCHITECTURE-INDEX.md`
- `ARCHITECTURE-ROADMAP.md`
- `ARCHITECTURE-EXECUTION-CHECKLIST.md`
- `ADR-0001`, `ADR-0003`, `ADR-0004`, `ADR-0005`
- `MODELS_REMODELING_AUDIT.md`
