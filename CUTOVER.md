# Cutover — virando o tráfego pro novo stack

As 7 fases da migração ([`docs/migracao/09-fases.md`](docs/migracao/09-fases.md))
estão com paridade de código completa: `frontend/` + `backend/` cobrem as 5 abas,
IA e o dashboard admin, e `index.html`/`firestore.rules` já saíram da raiz deste
repositório.

O que falta **não é código** — é um conjunto de passos que só quem tem acesso à
infraestrutura da ARCOM (hosting, DNS, contas dos provedores de API, projeto
Firebase) consegue executar. Esta lista é esse checklist.

## 1. Antes de virar o tráfego

- [ ] **Rotacionar as chaves** que estiveram no `index.html` público no GitHub
      Pages (não dá pra "des-vazar", só trocar): `ARCOM_API_KEY`,
      `TRACE360_API_KEY`, `GEOAPIFY_API_KEY`, `GROQ_API_KEY`, `TAVILY_API_KEY`.
      Guardar as novas com a infra, nunca no repositório — só em `.env` /
      variáveis de ambiente do host de produção.
- [ ] **Definir o host de produção do backend** (Go não roda no GitHub Pages).
      `docker compose up --build` já sobe frontend+backend+Postgres+Redis
      juntos — falta decidir onde esse compose roda em produção (servidor
      próprio, VM, o que a plataforma ARCOM já usa para outros serviços) e
      configurar TLS/domínio.
- [ ] **Exportar os dados vivos do Firestore** (usuários **aprovados**,
      `preCadastros`, `listasProspeccao`) via Firebase Admin SDK ou
      `gcloud firestore export`. Histórico de logs (`consultas_log`,
      `atividades_log`) **não precisa migrar** — decisão já tomada em
      [`docs/migracao/10-riscos-e-decisoes.md`](docs/migracao/10-riscos-e-decisoes.md#1-decisões-que-precisam-de-resposta),
      item 5: começa limpo no Postgres.
- [ ] **Converter o export pro formato JSON** que `backend/cmd/seed` espera
      (ver o comentário no topo de `backend/cmd/seed/main.go` — um array por
      coleção: `usuarios`, `preCadastros`, `listasProspeccao`).
- [ ] **Rodar o seed** contra o Postgres de produção:
      `go run ./cmd/seed -arquivo export.json`. Ele só migra contas com
      `status: aprovado` (pendentes/negadas do app antigo o usuário refaz o
      cadastro) e nunca preenche senha — ver o próximo item.
- [ ] **Avisar os usuários que vão precisar redefinir a senha** no primeiro
      acesso: o hash do Firebase Auth não é exportável, então ninguém migra
      com a senha antiga (decisão já tomada em
      [`docs/migracao/10-riscos-e-decisoes.md`](docs/migracao/10-riscos-e-decisoes.md#1-decisões-que-precisam-de-resposta),
      item 4). Ainda não existe fluxo de "esqueci minha senha" — no MVP, o
      admin recria a conta com uma senha temporária pelo painel
      `/admin` e passa pro usuário trocar em Configurações.
- [ ] **Validar as contagens** pós-seed antes de virar o tráfego (nº de
      usuários aprovados, pré-cadastros e listas batendo com o Firestore).

## 2. Virando o tráfego

- [ ] Apontar o domínio de produção pro novo deploy (Docker Compose /
      infraestrutura escolhida no item 1).
- [ ] Confirmar `GET /api/health` respondendo 200 no domínio de produção.
- [ ] Fazer um teste ponta a ponta em produção: login, consulta de CNPJ,
      prospecção, enriquecimento, IA e o dashboard admin — nessa ordem, com
      uma conta real aprovada.

## 3. Depois que o novo stack está estável

- [ ] **Aposentar o GitHub Pages**: desligar em Settings → Pages do
      repositório (o `index.html` que servia essa página já não existe mais
      no repositório).
- [ ] **Aposentar o projeto Firebase**: depois de confirmar que ninguém mais
      depende dele (nem o app antigo, que não tem mais onde rodar sem o
      `index.html`), desligar Authentication e Firestore, ou remover o
      projeto inteiro conforme a política de retenção da ARCOM.
- [ ] Atualizar qualquer link/bookmark interno que ainda aponte pro GitHub
      Pages antigo.

## 4. Fora do escopo deste repositório

Como registrado em [`docs/migracao/10-riscos-e-decisoes.md`](docs/migracao/10-riscos-e-decisoes.md#4-o-que-este-plano-não-cobre-fora-de-escopo):
provisionamento do ambiente de produção, custos dos provedores de API, e
qualquer funcionalidade nova além de paridade — isso entra depois do cutover.
