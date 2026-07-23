# 01 — Diagnóstico do app atual

Radiografia do `index.html` de hoje. Serve de checklist de paridade: nada aqui
pode "sumir" na migração sem decisão explícita.

## 1. Números

- 1 arquivo `index.html`, **~6.916 linhas / 366 KB**.
- **~250 funções** JS globais (sem módulos).
- 1 bloco `<style>` grande + 2 blocos `<script>` (linhas ~2321–6669 e ~6693–6913).
- 3 `<script>` de CDN do Firebase no `<head>`.
- Estado em `localStorage` (favoritos, histórico, tema, token Arcom, listas,
  chat, posse Trace360) e no **Firestore**.

## 2. As 5 abas (função `showTab`)

| Aba (`id`) | O que faz |
|------------|-----------|
| `tab-consulta` | Consulta CNPJ individual/lote (até 1000), valida dígito verificador antes de gastar chamada, cache 24h, favoritos com etiquetas, histórico, seletor de fonte (Arcom/BrasilAPI), mostra "Cliente Arcom" |
| `tab-prospeccao` | Busca empresas ativas não-clientes por cidade/UF/CNAE (multi-cidade/multi-ramo, dedup), corte de grandes redes por nº de filial, **score "provável loja física" 0–100** com semáforo, ranking por bairro, Street View embed, contatos clicáveis, pré-cadastro, PDF de indicação com rota |
| `tab-enriquecimento` | Dossiê **Trace360** por CNPJ; envio em lote (até 500), fluxo assíncrono com estados (na fila/processando/concluído/erro), posse por usuário guardada no navegador |
| `tab-conversao` | Controle de conversão (retorno da prospecção com assessor), processa **planilha de vendas** (xlsx/csv) para cruzar |
| `tab-admin` | Dashboard multiusuário: contas criadas/pendentes, logins, consultas, prospecções, PDFs, histórico de atividades filtrável, atividade por usuário, presença online; gráficos em Chart.js |

Recursos transversais: **Insight IA** e **chat IA** (Groq + Tavily),
exportação (Excel/CSV/JSON/PDF), tema claro/escuro + cor de destaque +
partículas, telas de cover/login/status, PWA-like.

## 3. Integrações externas (todas saem do navegador hoje)

### 3.1 API Consulta CNPJ Arcom — `https://consultacnpj.arcom.com.br`
- **Auth:** `POST /v1/auth/token` com `{ api_key }` → `{ token }` (vale ~90 dias,
  guardado em `localStorage.arcomToken`).
- **Consulta (individual e lote):** `POST /v1/cnpj/batch` com
  `{ cnpjs: [...], cliente: true }`, chunking de **até 1000** por chamada,
  retry em 401 (renova token). Resposta aninhada
  (`empresa.razao_social`, `municipio.descricao`, `situacao_cadastral`,
  `socios[]`, `simples.opcao_simples`, `ja_cliente`, `latitude/longitude`).
- **Municípios:** `GET /v1/municipios?q={cidade}` (resolve código do município).
- **Estabelecimentos (prospecção):** `GET /v1/estabelecimentos?uf=&municipioCodigo=&cnae=&offset=`
  (paginado).

### 3.2 BrasilAPI — `https://brasilapi.com.br/api/cnpj/v1/{cnpj}`
Pública, sem chave, 1 CNPJ por requisição. Fallback/alternativa da fonte Arcom.
Não informa "cliente Arcom" nem coordenadas.

### 3.3 Trace360 — `https://trace360ai.arcom.com.br/api/v1`
- `POST /clientes` com `[{cnpj}, ...]` → enfileira enriquecimento (retorna
  `resultados[]` com `cliente_id`).
- Endpoints de listagem/status por `cliente_id` (estados: pendente, enfileirado,
  processando, ambiguo_pausado, concluido, em_cache, nao_enriquecivel, erro,
  cancelado).
- Autenticação por header `x-api-key`.

### 3.4 Geoapify — `https://api.geoapify.com/v1/geocode/search`
Geocodificação de endereço (ordenar prospects por proximidade). Usa `apiKey` na
query string.

### 3.5 Groq + Tavily (via Cloudflare Worker)
- Worker `https://consultarcom.jlrodrigues6900.workers.dev` com rotas `/groq` e
  `/tavily` (as chaves reais ficam no Worker, **não** no bundle).
- **Groq:** `llama-3.3-70b-versatile` (insight/ranking) e `llama-3.1-8b-instant`
  (chat). Chat usa **tool-calling** com 2 ferramentas: `buscar_web` (→ Tavily) e
  `estatisticas_do_site` (lê números do próprio site em tempo real).
- **Tavily:** busca web para embasar o insight.

### 3.6 Google Maps / Street View
Embed por **endereço** (iframe), sem chave secreta. Usado no modal de "fachada".

### 3.7 Firebase — Auth + Firestore (SDK compat v10.13, via CDN)
Ver guia [`04`](04-modelo-dados.md) para as coleções.

## 4. Segredos expostos no bundle (crítico)

Em texto plano no `index.html` publicado (portanto **já vazados** para qualquer
visitante do GitHub Pages):

| Constante | Linha aprox. | Serviço | Situação |
|-----------|--------------|---------|----------|
| `TRACE360_API_KEY` | ~2447 | Trace360 (`x-api-key`) | **rotacionar** |
| `ARCOM_API_KEY` | ~3842 | Consulta CNPJ Arcom | **rotacionar** |
| `GEOAPIFY_API_KEY` | ~5855 | Geoapify | **rotacionar** |
| `firebaseConfig.apiKey` | ~3003 | Firebase | pública por design; sai com a migração |

> Groq/Tavily **não** estão no bundle (ficam no Worker) — mas o proxy migra para
> o backend Go de qualquer forma.

## 5. Regras de segurança de hoje

A autorização vive nas `firestore.rules` (super-admin por e-mail, `isAdmin`,
`isAprovado`, dono do doc). Na migração isso vira **autorização no backend Go** —
ver [`07`](07-seguranca.md). Não pode "cair" no meio do caminho.

## 6. Score "provável loja física" (regra de negócio a preservar)

Cálculo 100% client-side hoje (`scoreLojaFisica`), soma de sinais (0–100):

- **CNAE** (2 primeiros dígitos): 47 Varejo +40, 56 Alimentação +26, 45 Veículos
  +24, 86 Saúde +20, indústria +14, serviço/holding +0, outro +12.
- **Nome fantasia** presente +15.
- **Endereço com número** +15.
- **Telefone** +8 (fixo +6 extra).
- **Porte** acima de micro +8.
- **Coordenadas** presentes +5.
- **Tempo de atividade** ≥24 meses +8.

Vira serviço no backend (`internal/prospeccao`) — regra de negócio não fica
exposta no navegador (`padroes/07`).
