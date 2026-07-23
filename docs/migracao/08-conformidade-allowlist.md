# 08 — Conformidade com o allowlist

`padroes/01` é a fonte da verdade: **tudo fora dele é rejeitado no deploy**. O
CDA de hoje usa várias coisas fora do allowlist. Este guia resolve **cada uma**
pelo caminho que o padrão permite — sem inventar dependência.

## 1. O que sai e o substituto permitido

| Item hoje | Está no allowlist? | Substituto (dentro do padrão) |
|-----------|--------------------|-------------------------------|
| HTML/CSS/JS puro | ❌ | React + Vite + Tailwind + shadcn |
| Firebase Auth | ❌ | Go + `golang-jwt` + `bcrypt` + Postgres |
| Firestore | ❌ | Postgres (`gorm`/`pgx` + `goose`) + Redis (presença) |
| Font Awesome (CDN) | ❌ | `lucide-react` (já no stack) |
| JetBrains Mono (CDN) | ❌ (não é token da marca) | Red Hat Display (única fonte do DS) |
| **Chart.js** (CDN) | ❌ | §2 — gráficos em SVG/Tailwind à mão |
| **SheetJS (xlsx)** (CDN) | ❌ | §3 — CSV via `encoding/csv` no backend |
| **html2pdf.js** (CDN) | ❌ | §3 — HTML de impressão + print-to-PDF |
| **qrcodejs** (CDN) | ❌ | §3 — link de rota (QR opcional, ver §3) |
| `fetch` a domínio externo do navegador | ❌ proibido | proxy no backend (`/api/*`) |
| Cloudflare Worker (Groq/Tavily) | fora do modelo | endpoints `/api/ia/*` no backend |
| GitHub Pages | ❌ (não roda backend) | Docker na infra ([`10`](10-riscos-e-decisoes.md)) |

Libs que **já estão** no allowlist e continuam: `react-router-dom`,
`react-hook-form`, `axios`, `dayjs`, `tailwind*`, `@radix-ui/*`, `lucide-react`,
`vite-plugin-pwa`.

## 2. Gráficos do dashboard (era Chart.js)

Não há lib de gráfico no allowlist. Os gráficos do dashboard admin são simples
(contagens, séries no tempo, ranking) — **recriar com SVG + Tailwind** à mão:

- **Barras/ranking:** `<div>` com `width` proporcional + tokens de cor
  (`bg-verde-arcom`, `bg-verde-lima`). Não precisa de lib.
- **Séries no tempo:** `<svg>` com `<polyline>`/`<rect>` calculados no componente.
- Números-chave: cards de métrica (KPI) com os componentes shadcn.

> Se a infra futuramente liberar uma lib de gráfico no allowlist, dá pra trocar.
> Até lá, SVG/Tailwind cobre o dashboard sem sair do padrão.

## 3. Exportação (era SheetJS / html2pdf / qrcodejs)

Nenhuma dessas está no allowlist; o padrão manda gerar no backend/servidor:

- **CSV** (abre no Excel): `encoding/csv` da stdlib Go → `GET /api/.../exportar`
  devolve o arquivo. Cobre "Excel/CSV" com seleção de colunas.
- **JSON:** `encoding/json` (trivial).
- **XLSX nativo:** não há lib no allowlist. Opções: (a) entregar **CSV** como
  formato de planilha (recomendado no MVP); (b) montar XLSX à mão com
  `archive/zip` + `encoding/xml` da stdlib (mais trabalhoso); (c) pedir à infra a
  inclusão de uma lib Go de xlsx no allowlist. **Decisão em [`10`](10-riscos-e-decisoes.md).**
- **PDF de indicação:** backend renderiza um **HTML de impressão** com
  `html/template` (agrupado por ramo, com CNPJ/endereço/CEP/sócios/telefone e o
  link/QR da rota); o frontend abre e usa `window.print()` → "Salvar como PDF".
  Sem lib de PDF, dentro do padrão.
- **QR code da rota:** sem lib no allowlist. No MVP, o PDF traz o **link clicável**
  do Google Maps/Waze (o QR é conveniência). Se o QR for obrigatório, gerar SVG
  de QR na stdlib é possível porém trabalhoso — **decisão em [`10`]**.

## 4. Chamadas externas

Regra dura do padrão: o **navegador não chama domínio externo**. Toda integração
(Arcom, BrasilAPI, Trace360, Geoapify, Groq, Tavily) vira endpoint `/api` no
backend Go (ver [`03`](03-contrato-api.md)). Exceção: **embed** de Google
Street View/Maps por endereço (iframe, sem chave) — permanece no frontend por
não ser chamada de API com segredo.

## 5. Fontes e ícones

- **Fonte única:** Red Hat Display, via `@import` no `index.css` (permitido pelo
  DS). JetBrains Mono e Font Awesome saem.
- **Ícones:** `lucide-react`. **Sem emoji** em lugar nenhum da UI — inclusive o
  semáforo do score vira bolinhas com tokens de cor.

## 6. Resumo da decisão de conformidade

Tudo que hoje está fora do allowlist tem caminho **dentro** do padrão. As únicas
questões que dependem de terceiros (não de código) estão isoladas em
[`10`](10-riscos-e-decisoes.md): formato de planilha (CSV vs XLSX), QR
obrigatório ou não, e onde o backend vai rodar em produção.
