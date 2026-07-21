## Plan: Configurable Default API Version

Adicionar uma variÃ¡vel de ambiente para controlar qual versÃ£o responde no alias sem versÃ£o (`/api/...`), mantendo todas as versÃµes explÃ­citas funcionando em paralelo. A recomendaÃ§Ã£o Ã© separar o alias base do registro versionado, permitir `v4` e `v5` como defaults suportados para o alias sem versÃ£o, expor essa escolha em `settings` e `preview`, e migrar o frontend `src/apps/vuejs` para usar explicitamente `/api/v5/...` para evitar quebra quando `/api/...` apontar para `v4`.

**Steps**
1. Fase 1 â€” Modelar a nova configuraÃ§Ã£o de ambiente em `src/environment/api_settings.go`.
   - Adicionar constante da env, por exemplo `API_DEFAULT_VERSION`.
   - Adicionar campo em `environment.APISettings` para a versÃ£o default do alias sem versÃ£o.
   - Carregar com default inicial alinhado Ã  decisÃ£o de rollout (pelo seu cenÃ¡rio, `v4` durante a migraÃ§Ã£o; se a equipe preferir rollout neutro, manter `v5` e mudar apenas por env).
   - Validar/normalizar o valor para evitar aliases invÃ¡lidos. RecomendaÃ§Ã£o: suportar explicitamente `v4` e `v5` para o alias base; manter `v3` apenas como rota explicitamente versionada, porque seu shape (`/v3/bot/{token}`) nÃ£o combina com o contrato canÃ´nico sem versÃ£o.
2. Fase 2 â€” Expor a configuraÃ§Ã£o no environment discovery.
   - Atualizar `src/environment/environment_settings_preview.go` para incluir um campo pÃºblico como `default_api_version` no preview.
   - Garantir que o mesmo campo apareÃ§a em `settings` autenticado automaticamente via `APISettings`.
   - Revisar `src/api/api_handlers+EnvironmentController.go` apenas para validar que a sanitizaÃ§Ã£o continua correta e que nenhum segredo novo Ã© exposto.
3. Fase 3 â€” Desacoplar alias sem versÃ£o do registro fixo atual. *depende da Fase 1*
   - Hoje `src/api/v5/routes.go` sempre monta `""` e `"/v5"`, e `src/api/legacy/routes.go` sempre monta `""`, `"/current"` e `"/v4"`.
   - Refatorar isso para separar:
     - rotas explicitamente versionadas, sempre ativas;
     - alias sem versÃ£o, montado apenas para a versÃ£o escolhida por env.
   - A forma mais segura Ã© introduzir registradores mais granulares, por exemplo:
     - mount da famÃ­lia canÃ´nica somente em `"/v5"`;
     - mount da famÃ­lia legada somente em `"/v4"` e `"/current"`;
     - um mount adicional para `""` decidido por `environment.Settings.API.DefaultVersion`.
   - O ponto central dessa orquestraÃ§Ã£o deve ficar em `src/api/api.go`, porque ali jÃ¡ existe a decisÃ£o de montagem sob `API_PREFIX`.
4. Fase 4 â€” Preservar compatibilidade de rotas explÃ­citas. *pode ocorrer em paralelo com parte da Fase 3*
   - Manter sempre funcionando, independentemente do default:
     - `/api/v5/...` (canÃ´nica atual)
     - `/api/v4/...` e `/api/current/...` (legado atual)
     - `/api/v3/bot/{token}`
   - Manter tambÃ©m o comportamento existente de `API_PREFIX`, ou seja, a mudanÃ§a controla somente qual versÃ£o responde em `/<prefix>/...`, sem alterar o prefixo configurÃ¡vel.
5. Fase 5 â€” Migrar o frontend Vue para versÃ£o explÃ­cita. *depende do desenho de roteamento da Fase 3*
   - Atualizar `src/apps/vuejs/client/src/services/api.ts` para nÃ£o depender semanticamente do alias `/api/...` quando o objetivo for a API canÃ´nica nova.
   - RecomendaÃ§Ã£o: reescrever chamadas canÃ´nicas do SPA para `/api/v5/...` antes da substituiÃ§Ã£o por `apiBase`, preservando apenas a resoluÃ§Ã£o de prefixo configurado.
   - Validar se hÃ¡ componentes ou testes que assumem `/api/...` diretamente e ajustar o contrato do frontend para a v5 explÃ­cita.
6. Fase 6 â€” Atualizar testes de roteamento e discovery. *depende das Fases 2â€“5*
   - Ajustar `src/api/api_route_registration_test.go` para validar:
     - aliases explÃ­citos por versÃ£o continuam montados;
     - o alias sem versÃ£o muda conforme o env configurado;
     - `/api/v5/...` e `/api/v4/...` permanecem estÃ¡veis.
   - Atualizar `src/api/api_environment_discovery_test.go` e/ou `src/api/api_handlers+EnvironmentController_test.go` para verificar `default_api_version` em `preview` e `settings`.
   - Revisar `src/frontend_canonical_routes_test.go` e demais testes que assumem `/api/...` como v5 implÃ­cita.
   - Adicionar teste do loader de `APISettings` para o novo env e seu default.
7. Fase 7 â€” Atualizar documentaÃ§Ã£o operacional. *depende da implementaÃ§Ã£o fechada*
   - Documentar a nova env em `src/environment/README.md`.
   - Atualizar `docs/USAGE-environment-discovery.md` com o novo campo no payload.
   - Atualizar `README.md` e, se necessÃ¡rio, `docs/USAGE-authentication-modes.md` para deixar claro que:
     - `/api/v4/...` e `/api/v5/...` coexistem;
     - `/api/...` aponta para a versÃ£o definida por env;
     - o SPA oficial usa `v5` explÃ­cita para nÃ£o ser afetado pelo alias default.
8. Fase 8 â€” VerificaÃ§Ã£o final.
   - Rodar testes focados de API e roteamento.
   - Validar manualmente exemplos como:
     - `/api/system/version` respondendo pela versÃ£o default configurada;
     - `/api/v4/health` permanecendo funcional;
     - `/api/v5/system/version` permanecendo funcional;
     - `/api/system/environment` mostrando `default_api_version` em `preview` e `settings`.

**Relevant files**
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\environment\api_settings.go` â€” declarar/carregar `API_DEFAULT_VERSION` em `APISettings`.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\environment\environment_settings_preview.go` â€” expor `default_api_version` no preview pÃºblico.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\api.go` â€” ponto principal para decidir qual versÃ£o recebe o alias sem versÃ£o sob `API_PREFIX`.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\api_handlers_v5.go` â€” hoje registra a famÃ­lia canÃ´nica com alias vazio; provÃ¡vel refatoraÃ§Ã£o para mount explÃ­cito e/ou parametrizado.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\v5\routes.go` â€” separar mount de alias vazio e mount de alias versionado (`/v5`).
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\api_handlers.go` â€” famÃ­lia legada atual (`v4`) e possÃ­vel ponto para expor mount explÃ­cito.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\legacy\routes.go` â€” hoje monta `""`, `"/current"`, `"/v4"`; precisa desacoplar o alias vazio.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\apps\vuejs\client\src\services\api.ts` â€” migrar SPA para `v5` explÃ­cita.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\api_route_registration_test.go` â€” cobertura de aliases/versionamento.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\api_environment_discovery_test.go` â€” cobertura de `default_api_version` no endpoint de environment.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\api_handlers+EnvironmentController_test.go` â€” asserts adicionais de discovery, se necessÃ¡rio.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\frontend_canonical_routes_test.go` â€” revisar expectativas do frontend sobre `/api/...`.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\environment\README.md` â€” documentar a nova env.
- `z:\Desenvolvimento\nocodeleaks-quepasa\docs\USAGE-environment-discovery.md` â€” documentar o novo campo exposto.
- `z:\Desenvolvimento\nocodeleaks-quepasa\README.md` â€” documentar comportamento do alias default da API.

**Verification**
1. Rodar testes de API conforme a convenÃ§Ã£o do repositÃ³rio: `cd src/api` e executar testes focados de registro de rota, environment discovery e controllers relacionados.
2. Rodar build do backend em `src` para validar que a reorganizaÃ§Ã£o do registro de rotas nÃ£o quebrou a montagem global.
3. Validar manualmente, com a env apontando para `v4`, que:
   - `/api/system/version` e demais endpoints sem versÃ£o caem no comportamento legado esperado;
   - `/api/v5/system/version` continua acessÃ­vel;
   - `/api/v4/health` continua acessÃ­vel.
4. Validar manualmente, com a env apontando para `v5`, que o comportamento volta ao padrÃ£o canÃ´nico atual.
5. Validar o SPA Vue apÃ³s a migraÃ§Ã£o para `/api/v5/...`, confirmando que ele continua funcional independentemente do valor de `API_DEFAULT_VERSION`.
6. Se houver alteraÃ§Ã£o em payload documentado ou anotaÃ§Ãµes expostas, regenerar/validar Swagger de acordo com a convenÃ§Ã£o do repositÃ³rio.

**Decisions**
- Confirmado: o SPA oficial deve ser migrado para usar `v5` explÃ­cita, evitando conflitos com o alias `/api/...`.
- Confirmado: a versÃ£o default da API deve aparecer tanto em `settings` quanto em `preview` do endpoint de environment.
- RecomendaÃ§Ã£o arquitetural: limitar o alias sem versÃ£o configurÃ¡vel a `v4` e `v5`; manter `v3` somente como rota explicitamente versionada.
- IncluÃ­do no escopo: coexistÃªncia total das rotas explicitamente versionadas.
- ExcluÃ­do do escopo: remover versÃµes antigas, mudar `API_PREFIX`, ou alterar o contrato explÃ­cito de `/api/v3/...`, `/api/v4/...`, `/api/v5/...`.

**Further Considerations**
1. Valor default da nova env: para atender ao rollout de migraÃ§Ã£o, a recomendaÃ§Ã£o Ã© usar `v4` como default inicial em produÃ§Ã£o, mas isso deve ser decisÃ£o consciente porque muda o comportamento histÃ³rico atual de `/api/...`.
2. Nomenclatura: `API_DEFAULT_VERSION` comunica melhor a intenÃ§Ã£o do que algo ligado a â€œcanonicalâ€, porque o alias pode apontar temporariamente para a famÃ­lia legada.
3. Se a equipe quiser endurecer a configuraÃ§Ã£o, Ã© recomendÃ¡vel tratar valores invÃ¡lidos com fallback claro e log explÃ­cito no startup para facilitar suporte operacional.
