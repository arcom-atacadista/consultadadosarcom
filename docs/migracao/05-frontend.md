# 05 — Frontend (React + Vite + Tailwind + shadcn)

Aplica `padroes/02` e `padroes/08`. O frontend **só desenha e chama `/api`** —
nenhuma lógica de negócio, nenhum segredo.

## 1. Estrutura `src/`

```
src/
├── main.tsx                 # entrada + router (react-router-dom)
├── App.tsx                  # shell: sidebar + topbar + <Outlet/>
├── index.css                # @import Red Hat Display + @tailwind
├── lib/
│   ├── api.ts               # axios baseURL "/api" + interceptor de token
│   └── cn.ts                # helper de classes (clsx + tailwind-merge)
├── components/
│   ├── ui/                  # shadcn com tokens ARCOM (Button, Card, Badge,
│   │                        #   Input, Select, Alert, Dialog, Table, Tabs...)
│   ├── layout/              # Sidebar, Topbar, NavItem
│   ├── cnpj/                # CardEmpresa, ListaSocios, SeletorFonte
│   ├── prospeccao/          # TabelaProspects, ScoreSemaforo, ModalFachada
│   ├── ia/                  # PainelInsight, ChatIA
│   └── admin/               # cards de métrica, GraficoBarras/Linhas (SVG)
├── pages/
│   ├── Login.tsx
│   ├── Cover.tsx            # tela inicial
│   ├── Consulta.tsx
│   ├── Prospeccao.tsx
│   ├── Enriquecimento.tsx
│   ├── Conversao.tsx
│   ├── Admin.tsx
│   └── Pendente.tsx         # conta aguardando aprovação
└── hooks/
    ├── useAuth.ts           # token, usuário, guarda de rota
    ├── useConsulta.ts
    ├── useProspeccao.ts
    └── usePresenca.ts       # heartbeat para /api/presenca
```

## 2. Comunicação com o backend — sempre `/api`

```ts
// src/lib/api.ts
import axios from "axios";
export const api = axios.create({ baseURL: "/api" });
api.interceptors.request.use((cfg) => {
  const t = localStorage.getItem("token");
  if (t) cfg.headers.Authorization = `Bearer ${t}`;
  return cfg;
});
```

Nunca `http://...` na mão. Nada de `fetch` para domínio externo — o backend faz.
Só o **token JWT** fica no navegador (não é segredo de terceiro).

## 3. Navegação (react-router-dom)

Rotas num só lugar (`main.tsx`). Guarda por estado do usuário:

| Rota | Página | Acesso |
|------|--------|--------|
| `/` | Cover | público |
| `/login` | Login | público |
| `/pendente` | Pendente | logado, não aprovado |
| `/consulta` | Consulta | aprovado |
| `/prospeccao` | Prospeccao | aprovado |
| `/enriquecimento` | Enriquecimento | aprovado |
| `/conversao` | Conversao | aprovado |
| `/admin` | Admin | admin |

## 4. Design System ARCOM (obrigatório — padroes/08)

- **Fonte:** Red Hat Display via `@import` no topo do `index.css`, antes das
  diretivas `@tailwind`. (JetBrains Mono do app antigo **sai** — não é token da
  marca.)
- **Cores só dos tokens** do `tailwind.config.js`: `verde-escuro`, `verde-arcom`,
  `verde-lima`, `arcom-gray`, `danger`, `surface`, `surface-border`.
- **Fundo da página** `bg-surface` (nunca branco puro). Verde é **acento**.
- **Sidebar** `bg-verde-escuro`, nav ativo com friso `border-l-[3px]
  border-verde-lima`. **Topbar** branca com borda inferior `surface-border`.
- **Cards** brancos, `border-surface-border`, `shadow-sm`; hover
  `-translate-y-0.5 shadow-md`.
- **Ícones** `lucide-react` (substituem Font Awesome), stroke 2px. **Sem emoji.**
  → O semáforo do score de loja física (hoje 🟢/🟡/🔴) vira **bolinhas coloridas
  com os tokens** (`verde-arcom`/`verde-lima`/`danger`), sem emoji.
- **Tom de voz:** direto, profissional; dinheiro `R$ 1.250,00`; quantidades
  `24un`/`12cx`. Marca **ARCOM**.

## 5. Componentes shadcn a montar (com tokens)

Button (primary/secondary/ghost/danger/dark/accent), Card
(default/brand/accent/outlined), Badge/Tag (pill), Input/Select (label/hint/erro),
Alert (success/danger/warning/info/brand), Dialog (modais), Table (densidade B2B),
Tabs (as 5 abas). Variantes conforme `padroes/08`.

## 6. Estado e libs (só allowlist)

- **Formulários:** `react-hook-form` (login, filtros, pré-cadastro).
- **Datas:** `dayjs` (+ `utc`/`timezone`). `moment` proibido.
- **HTTP:** `axios` (`/api`).
- **PWA:** `vite-plugin-pwa` — no bloco `manifest` ajustar nome "CDA — Consulta
  Dados Arcom", `short_name` "CDA", `theme_color` `#007840`.
- **Preferências locais** (tema, cor de destaque, densidade, favoritos, histórico
  de consulta): podem seguir em `localStorage` (não são segredo nem dado
  compartilhado). Favoritos/histórico **poderão** migrar para o backend depois se
  precisarem sincronizar entre aparelhos (sinal de persistência — `padroes/07`).

## 7. Gráficos, exportação, mapa

- **Gráficos** (dashboard): SVG/Tailwind à mão — ver [`08`](08-conformidade-allowlist.md).
- **Exportação:** backend gera; frontend só baixa/aciona print — ver [`08`].
- **Street View/Maps:** iframe por endereço, sem chave (segue no frontend).
