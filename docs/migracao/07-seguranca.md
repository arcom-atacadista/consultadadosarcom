# 07 — Segurança, segredos e autorização

O maior ganho da migração é de segurança: **tirar os segredos do navegador** e
mover a autorização (hoje nas `firestore.rules`) para o backend Go.

## 1. Segredos — ação imediata

As três chaves abaixo estão **publicadas** no `index.html` do GitHub Pages
(qualquer um lê). Não dá para "desvazar" — precisa **rotacionar**:

| Chave | Serviço | O que fazer |
|-------|---------|-------------|
| `ARCOM_API_KEY` | Consulta CNPJ Arcom | rotacionar; nova chave só no ambiente do backend |
| `TRACE360_API_KEY` | Trace360 | rotacionar; só no backend |
| `GEOAPIFY_API_KEY` | Geoapify | rotacionar; só no backend |
| Groq / Tavily | via Worker | migrar do Worker para env do backend |
| `firebaseConfig.apiKey` | Firebase | sai com a migração (deixa de existir) |

Regra do padrão (`padroes/01`, `05`): **nenhum segredo no bundle**; só `VITE_*`
públicas. No CDA, isso significa **zero** chave de terceiro no frontend — todas
passam a ser variáveis de ambiente do serviço `backend` (ver [`02`](02-arquitetura-alvo.md) §3).

## 2. Autenticação

- **Login:** `POST /api/auth/login` valida e-mail/senha (bcrypt) e devolve **JWT**
  assinado com `JWT_SECRET`. Claims: `uid`, `email`, `nome`, `isAdmin`, `aprovado`
  — os dois últimos já vêm calculados (super-admin por e-mail OU coluna do
  banco), então o middleware só lê o token, sem round-trip no Postgres por
  requisição. Expira em 24h.
- **Cadastro:** `POST /api/auth/register` cria usuário `pendente`, `isAdmin=false`
  (idêntico à regra Firestore de hoje).
- **Token** no navegador em `localStorage.token`; enviado em `Authorization:
  Bearer`. Expiração curta + renovação (a definir; hoje o Firebase gerenciava).

## 3. Autorização — de `firestore.rules` para middleware Go

As regras atuais têm três níveis. Tradução direta:

| Regra Firestore (hoje) | Middleware Go |
|------------------------|---------------|
| `emailAdmin()` (super-admin por e-mail) | `SUPER_ADMIN_EMAIL` no config |
| `isAdmin()` (super-admin OU `isAdmin==true`) | `RequireAdmin` |
| `isAprovado()` (admin OU `status=='aprovado'`) | `RequireAprovado` |
| `isOwner(uid)` (dono do doc) | checagem por `uid` do JWT no service |

Mapeamento das permissões por coleção → rota:

| Coleção / regra hoje | Rota / regra no backend |
|----------------------|-------------------------|
| `usuarios`: lê o próprio; admin lê todos; só admin edita | `GET /api/auth/me` (próprio); `/api/usuarios*` só admin |
| `consultas_log`: create/read aprovado; edita admin | log via middleware; leitura em `/api/admin/*` |
| `atividades_log`: create logado; read aprovado; edita admin | idem |
| `presenca`: read admin; escreve o dono | `POST /api/presenca` (dono); `GET /api/admin/presenca` |
| `preCadastros`: aprovado faz tudo | `/api/precadastros*` (RequireAprovado) |
| `listasProspeccao`: dono lê a própria; admin tudo; edita admin | `/api/prospeccao/listas*` com checagem de dono |
| `prospeccoes`: create aprovado; read/edit admin | grava no `buscar`; leitura admin |
| fallback `deny all` | rota inexistente → 404; sem grupo → 401/403 |

**Importante:** a segurança hoje está **no cliente confiando nas regras do
Firestore**. No alvo, o backend é a única autoridade — validar sempre no
servidor, nunca confiar no frontend (`padroes/03`, `05`).

## 4. Boas práticas herdadas do padrão

- Go **não roda script de instalação** de dependência (`padroes/01`): o vetor de
  "postinstall malicioso" do npm não existe no backend.
- Frontend com `.npmrc` `ignore-scripts=true` (não alterar) fecha o mesmo vetor
  no lado do build.
- `go.sum` e `pnpm-lock.yaml` commitados garantem integridade das dependências.
- Rate limit / limites de terceiros (429) tratados no backend com mensagem clara.

## 5. Dados sensíveis em trânsito

- Só o frontend publica porta; backend/banco/redis ficam na rede interna do
  compose (`padroes/04`).
- Em produção, TLS/domínio ficam a cargo da infra da plataforma (ver
  [`10`](10-riscos-e-decisoes.md)).
