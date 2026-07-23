# 03 — Contrato da API (`/api`)

Todo endpoint fica sob `/api` (padroes/03). Autenticação por **JWT** no header
`Authorization: Bearer <token>`, exceto onde indicado. Corpo em JSON. Validação
com `go-playground/validator`. Datas em ISO-8601 (UTC).

Convenções de erro:

```json
{ "erro": "mensagem curta e clara em português" }
```

Códigos: `400` validação, `401` sem/token inválido, `403` sem permissão,
`404` não encontrado, `409` conflito, `429` limite de terceiro, `502` falha na
API externa.

---

## 1. Saúde
| Método | Rota | Auth | Descrição |
|--------|------|------|-----------|
| GET | `/api/health` | não | `{ "status": "ok" }` (200) — usado no healthcheck |

## 2. Autenticação e usuários (`internal/auth`)
| Método | Rota | Auth | Descrição |
|--------|------|------|-----------|
| POST | `/api/auth/register` | não | cria usuário **pendente** (`{ email, senha, nome }`) |
| POST | `/api/auth/login` | não | `{ email, senha }` → `{ token, usuario }` |
| GET | `/api/auth/me` | sim | perfil do usuário logado |
| POST | `/api/auth/senha` | sim | troca de senha `{ atual, nova }` |
| GET | `/api/usuarios` | admin | lista usuários (status/admin) |
| PATCH | `/api/usuarios/:id` | admin | aprovar/negar/promover `{ status, isAdmin }` |
| DELETE | `/api/usuarios/:id` | admin | remove usuário |

`login` sempre devolve token se a senha bater — autenticação e autorização são
coisas separadas. O token carrega `isAdmin`/`aprovado` já calculados (ver
`efetivos()` em `07-seguranca.md`); rotas `aprovado`/`admin` é que barram quem
não tem o nível certo. Isso deixa o front levar quem está pendente pra tela de
espera com um token válido, em vez de travar no próprio login. Cadastro sempre
`pendente` e `isAdmin=false` (igual às regras Firestore de hoje).

Como o JWT é stateless, uma aprovação feita por um admin só passa a valer no
**próximo login** do usuário (o token antigo continua com `aprovado=false` até
expirar ou ele logar de novo).

## 3. CNPJ (`internal/cnpj`)
| Método | Rota | Auth | Descrição |
|--------|------|------|-----------|
| POST | `/api/cnpj/consultar` | aprovado | `{ cnpjs: [...], fonte: "arcom"\|"brasilapi" }` → lista normalizada; grava `consultas_log` |

- Backend guarda o **token da API Arcom** (troca via `POST /v1/auth/token` com a
  `ARCOM_API_KEY` do ambiente) — o cliente **nunca** vê a chave.
- Lote com chunking de até 1000 (fonte Arcom); fallback BrasilAPI 1 a 1.
- **Cache 24h** por CNPJ (Redis) — reconsulta não gasta chamada.
- Resposta normalizada num formato único (o mesmo que o front já espera):
  razão, fantasia, situação, sócios[], endereço, telefone, `clienteArcom`,
  `latitude/longitude`, `api` (origem).

## 4. Prospecção (`internal/prospeccao`)
| Método | Rota | Auth | Descrição |
|--------|------|------|-----------|
| GET | `/api/prospeccao/municipios?q=` | aprovado | resolve código de município |
| POST | `/api/prospeccao/buscar` | aprovado | `{ ufs[], cidades[], cnaes[], cortarRedes }` → estabelecimentos com **score de loja física** já calculado; registra em `prospeccoes` |
| GET | `/api/prospeccao/listas` | aprovado | listas salvas do usuário |
| POST | `/api/prospeccao/listas` | aprovado | salva lista `{ nome, filtros, itens }` |
| DELETE | `/api/prospeccao/listas/:id` | dono/admin | apaga lista |

- Paginação server-side sobre `/v1/estabelecimentos`, dedup multi-cidade/ramo.
- **Corte de grandes redes** e **score 0–100** (ver [`01`](01-diagnostico-atual.md) §6)
  calculados no backend.

## 5. Pré-cadastro (`internal/prospeccao`)
| Método | Rota | Auth | Descrição |
|--------|------|------|-----------|
| GET | `/api/precadastros` | aprovado | lista (compartilhada) |
| POST | `/api/precadastros` | aprovado | cria `{ cnpj, razao, endereco, contato, notas, status }` |
| PATCH | `/api/precadastros/:id` | aprovado | edita status/notas |
| DELETE | `/api/precadastros/:id` | aprovado | remove |

## 6. Geo (`internal/geo`)
| Método | Rota | Auth | Descrição |
|--------|------|------|-----------|
| GET | `/api/geo/geocode?q=` | aprovado | proxy Geoapify; chave no backend |

## 7. Enriquecimento Trace360 (`internal/enriquecimento`)
| Método | Rota | Auth | Descrição |
|--------|------|------|-----------|
| POST | `/api/enriquecimento` | aprovado | `{ cnpjs: [...] }` (até 500) → enfileira na Trace360; grava posse por usuário |
| GET | `/api/enriquecimento` | aprovado | lista os enviados **pelo usuário** com status atual |
| GET | `/api/enriquecimento/:clienteId` | dono | dossiê/status detalhado |

- Chave `x-api-key` no backend. **Posse por usuário** deixa de ser localStorage
  e passa a tabela (`enriquecimentos`, ver [`04`](04-modelo-dados.md)).
- Atualização de status via `asynq`/`cron` (polling na Trace360) ou sob demanda.

## 8. IA (`internal/ia`)
| Método | Rota | Auth | Descrição |
|--------|------|------|-----------|
| POST | `/api/ia/insight` | aprovado | insight comercial de 1 empresa (Groq 70B + Tavily) |
| POST | `/api/ia/ranking` | aprovado | ranking de leads por IA (lote) |
| POST | `/api/ia/chat` | aprovado | chat (Groq 8B) com **tool-calling** |

Ferramentas do chat resolvidas **no backend**:
- `buscar_web(query)` → Tavily (backend).
- `estatisticas_do_site()` → números reais do backend (presença, contas,
  consultas, prospecções, PDFs, ranking) — **não** expõe dados a quem não é
  admin; o backend filtra por permissão.

## 9. Conversão (`internal/conversao`)
| Método | Rota | Auth | Descrição |
|--------|------|------|-----------|
| GET | `/api/conversao` | aprovado | retorno da prospecção (com assessor) |
| POST | `/api/conversao/planilha` | aprovado | envia planilha de vendas (multipart) → cruza e atualiza |
| GET | `/api/conversao/exportar` | aprovado | exportação (ver [`08`](08-conformidade-allowlist.md)) |

## 10. Admin (`internal/admin`)
| Método | Rota | Auth | Descrição |
|--------|------|------|-----------|
| GET | `/api/admin/dashboard` | admin | agregados: contas, logins, consultas, prospecções, PDFs |
| GET | `/api/admin/atividades?tipo=` | admin | histórico filtrável |
| GET | `/api/admin/atividades/por-usuario` | admin | tabela por pessoa |
| GET | `/api/admin/presenca` | admin | quem está online (Redis) |
| POST | `/api/presenca` | aprovado | heartbeat do próprio usuário |

## 11. Log de atividades (transversal)

Middleware do backend registra `atividades_log` e `consultas_log` (login,
consulta, prospecção, PDF, etc.) — hoje o cliente escreve direto no Firestore;
passa a ser responsabilidade do servidor, com o nome/uid do usuário do JWT.
