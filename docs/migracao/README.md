# Documentação da Migração — CDA → Padrão ARCOM AI-First

> Migração do **CDA (Consulta Dados Arcom)** — hoje um único `index.html` no
> GitHub Pages — para a arquitetura AI-First da ARCOM (v2.1): monorepo
> `frontend/` (React + Vite + Tailwind + shadcn) + `backend/` (Go + chi),
> subindo com `docker compose up --build`.
>
> **Fonte da verdade do padrão:** a spec `arcom-projeto/` (pasta `padroes/`).
> O padrão vence qualquer preferência. Esta documentação apenas aplica o padrão
> ao caso concreto do CDA.

## Como ler (guias, um por assunto)

| Guia | Assunto |
|------|---------|
| [`00-visao-geral.md`](00-visao-geral.md) | Resumo executivo, princípios, o que muda e por quê |
| [`01-diagnostico-atual.md`](01-diagnostico-atual.md) | O app de hoje em detalhe: features, stack, endpoints, segredos |
| [`02-arquitetura-alvo.md`](02-arquitetura-alvo.md) | Arquitetura alvo: monorepo, fluxo de rede, Docker |
| [`03-contrato-api.md`](03-contrato-api.md) | Todos os endpoints `/api` (request/response) |
| [`04-modelo-dados.md`](04-modelo-dados.md) | Firestore → Postgres (tabelas, migrations) + Redis |
| [`05-frontend.md`](05-frontend.md) | Estrutura React, páginas, componentes shadcn, Design System |
| [`06-backend.md`](06-backend.md) | Pacotes Go em detalhe |
| [`07-seguranca.md`](07-seguranca.md) | Segredos, auth JWT/bcrypt, autorização (regras Firestore → Go) |
| [`08-conformidade-allowlist.md`](08-conformidade-allowlist.md) | O que sai da stack e o substituto permitido de cada coisa |
| [`09-fases.md`](09-fases.md) | Fases da migração com tarefas e critérios de aceite |
| [`10-riscos-e-decisoes.md`](10-riscos-e-decisoes.md) | Riscos e decisões que dependem de terceiros (infra/chaves) |

## Regra de ouro

Navegador → só o **frontend** (`:8080`). Toda chamada externa e todo segredo
ficam no **backend Go**, acessados via `/api/*` (proxy do Vite, sem CORS).
**Nenhum segredo no bundle do frontend.**
