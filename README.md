# CDA — Consulta Dados Arcom

Aplicação web para consulta, prospecção e enriquecimento comercial de empresas via CNPJ. Feita para uso interno da Arcom Atacadista.

Migrado do `index.html` monolítico original (HTML/CSS/JS puro + Firebase) para o padrão ARCOM AI-First: `frontend/` (React + Vite + Tailwind + shadcn) + `backend/` (Go + chi + Postgres + Redis), orquestrados via Docker Compose. As 7 fases da migração estão documentadas em [`docs/migracao/`](docs/migracao/README.md) — diagnóstico do app antigo, arquitetura alvo, contrato de API, modelo de dados, conformidade com o allowlist e o plano fase a fase.

**Antes de virar o tráfego de produção**, veja [`CUTOVER.md`](CUTOVER.md): a lista de passos manuais (rotação de chaves, export/seed de dados, DNS, aposentadoria do Firebase/GitHub Pages) que dependem de acesso à infraestrutura e não são feitos por este repositório sozinho.

---

## 🚀 Rodar localmente

```bash
docker compose up --build   # abre em http://localhost:8080
```

`GET /api/health` responde `{"status":"ok"}`. Sem Docker (dois terminais):

```bash
# backend
cd backend && cp .env.example .env && go run ./cmd/server

# frontend
cd frontend && pnpm install && pnpm run dev
```

Variáveis de ambiente: `.env.example` (raiz, pro `docker compose`) e `backend/.env.example` (backend fora do Docker). Nenhuma chave de API externa fica no navegador — tudo passa pelo backend (`/api/*`).

---

## ✨ Funcionalidades

### Consulta de CNPJ
- Consulta individual ou em lote (cole vários CNPJs, um por linha).
- Fonte de dados: **Consulta CNPJ Arcom** (API própria, chunk de até 1000 CNPJs) ou **Brasil API** como alternativa.
- Mostra se a empresa já é cliente Arcom, com todos os campos que a API retorna.
- Cache de 24h no Redis: reconsultar o mesmo CNPJ não gasta uma nova chamada.
- Favoritos e histórico das últimas consultas (localStorage).

### Prospecção
- Busca de empresas ativas (e que ainda não são clientes Arcom) por cidade, UF e ramo (CNAE), múltiplas cidades/ramos numa busca.
- Corte de grandes redes e score de "provável loja física" (0-100, com sinais explicados), calculado no backend.
- Modal Fachada/Mapa com Street View embutido, contatos clicáveis, pré-cadastro direto.
- Listas de prospecção salvas com nome e assessor, pra alimentar o relatório de Conversão.
- Indicação de clientes em PDF (impressão do navegador, agrupado por ramo, com rota de visita).

### Enriquecimento (Trace360)
- Dossiê completo por CNPJ via Trace360, em lote (até 500), com acompanhamento de status/progresso, download do PDF e dos dados.

### IA
- Insight comercial por empresa (Groq + busca na web via Tavily), ranking de leads por IA e chat com histórico, tudo com as chaves só no backend.

### Conversão da prospecção
- De cada lista salva pra um assessor, quanto virou Cliente Arcom — por assessor, por quem prospectou e por cidade, com exportação em CSV.

### Administração
- Login com e-mail/senha (JWT + bcrypt); contas novas ficam pendentes até um admin aprovar (ou negar).
- Dashboard com contas, presença online (Redis), histórico de atividades filtrável e atividade por usuário.

### Interface
- Design system ARCOM (Red Hat Display, tokens de cor fixos, `lucide-react`, sem emoji).
- PWA instalável, acessibilidade (`prefers-reduced-motion`).

---

## 🧱 Stack técnica

| Camada | Tecnologia |
|---|---|
| Frontend | React + Vite + Tailwind + shadcn (`react-router-dom`, `axios`, `react-hook-form`) |
| Backend | Go + chi + gorm (Postgres) + go-redis |
| Autenticação | JWT (`golang-jwt`) + bcrypt, contas em Postgres |
| Banco | Postgres (dados persistentes) + Redis (cache de CNPJ, presença) |
| Gráficos | SVG/Tailwind à mão (sem lib de gráfico) |
| Exportação | `encoding/csv` (backend) + impressão do navegador pra PDF |
| Mapa/validação | Google Maps + Street View (embed, por endereço) |
| IA | Groq (chat completions) + Tavily (busca na web), chamados só pelo backend |
| Dados de CNPJ | API própria "Consulta CNPJ Arcom", Brasil API |
| Enriquecimento | Trace360 |
| Deploy | Docker Compose |

Detalhes de cada decisão (por que Postgres em vez de Firestore, por que SVG em vez de Chart.js, etc.) estão em [`docs/migracao/08-conformidade-allowlist.md`](docs/migracao/08-conformidade-allowlist.md).

---

## 📄 Licença

Uso interno — Arcom Atacadista.
