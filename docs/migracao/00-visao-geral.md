# 00 — Visão geral

## O problema em uma frase

O CDA hoje é um **único `index.html`** (~6.900 linhas, 366 KB) de HTML/CSS/JS
puro que roda no navegador e fala **direto** com Firebase e várias APIs — com
**três chaves secretas em texto plano no código**. O padrão ARCOM exige que o
frontend só desenhe a tela e chame `/api/...`, e que **toda lógica, segredo e
dado** fiquem num backend Go. Migrar é, portanto, **reconstruir de forma
estruturada** — e isso resolve de quebra a exposição de segredos.

## Princípios (vêm do padrão ARCOM)

1. **Frontend burro, backend dono.** O React desenha e chama `/api`; o Go tem a
   lógica, os segredos e os dados. (`padroes/00`, `03`)
2. **Uma origem só, sem CORS.** O Vite faz proxy de `/api` pro backend. O
   navegador nunca chama domínio externo direto. (`padroes/02`)
3. **Só o allowlist.** Nada fora de `padroes/01-stack-permitida.md`. O que a
   stack atual usa e não está no allowlist tem substituto definido no guia
   [`08`](08-conformidade-allowlist.md).
4. **Nenhum segredo no navegador.** Só variáveis `VITE_*` públicas. (`padroes/05`)
5. **Design System ARCOM em toda UI.** Cores/fonte/raios/sombras só dos tokens
   do `tailwind.config.js`; sem emoji. (`padroes/08`)
6. **Frontend primeiro, mas aqui já sabemos que precisa de backend** — há login,
   segredos e APIs externas (`padroes/07`).
7. **Sobe com um comando:** `docker compose up --build` → `http://localhost:8080`.

## O que muda (visão de topo)

| Dimensão | Hoje | Alvo |
|----------|------|------|
| Frontend | HTML/CSS/JS puro, sem build | React + Vite + Tailwind + shadcn (pnpm) |
| Backend | não existe (browser faz tudo) | Go + chi, tudo sob `/api` |
| Auth | Firebase Authentication | Go + JWT + bcrypt (Postgres) |
| Dados | Firestore (7 coleções) | PostgreSQL (`gorm`/`pgx` + `goose`) |
| Presença/cache/fila | Firestore / localStorage | Redis (`go-redis` + `asynq`) |
| Segredos | **hardcoded no bundle** | variáveis de ambiente no backend |
| Chamadas externas | direto do navegador | proxied pelo backend Go |
| Ícones | Font Awesome (CDN) | `lucide-react` |
| Gráficos | Chart.js (CDN) | SVG/Tailwind à mão (ver [`08`](08-conformidade-allowlist.md)) |
| Export | SheetJS/html2pdf/qrcode (CDN) | gerado no backend / print-to-PDF (ver [`08`](08-conformidade-allowlist.md)) |
| Deploy | GitHub Pages (estático) | Docker na infra da plataforma |

## O que NÃO muda (paridade de produto)

As 5 áreas e os recursos de IA continuam existindo — o usuário final não perde
função. Muda **onde** cada coisa roda, não **o que** o produto faz:

- **Consulta** de CNPJ (individual/lote, cache, favoritos, "cliente Arcom")
- **Prospecção** (busca por cidade/UF/CNAE, corte de redes, score de loja
  física + Street View, PDF de indicação)
- **Enriquecimento** (dossiê Trace360)
- **Conversão** (retorno da prospecção + planilha de vendas)
- **Admin** (dashboard multiusuário, atividades, presença)
- **IA** transversal (Insight/ranking com Groq + Tavily; chat com tool-calling)

## Ponto de partida do repositório

Hoje a raiz tem só `index.html`, `README.md` e `firestore.rules`. A migração
introduz `frontend/`, `backend/`, `docker-compose.yml` e `.env.example` como
descrito no guia [`02`](02-arquitetura-alvo.md), preservando o `index.html`
atual como **referência de paridade** até o novo app cobrir tudo.
