# CDA — Consulta Dados Arcom

Aplicação web (single-page, HTML/CSS/JS puro) para consulta, prospecção e enriquecimento comercial de empresas via CNPJ. Feita para uso interno da Arcom Atacadista.

🔗 Deploy: https://arcom-atacadista.github.io/consultadadosarcom/

---

## ✨ Funcionalidades

### Consulta de CNPJ
- Consulta individual ou em lote (cole vários CNPJs, um por linha).
- 3 fontes de dados à escolha: **BrasilAPI**, **OpenCNPJ** e **Consulta CNPJ Arcom** (API própria, com autenticação JWT e suporte a lote de até 1000 CNPJs por chamada).
- Cache local de 24h: reconsultar o mesmo CNPJ na mesma API não gasta uma nova chamada.
- Validação de CNPJ no cliente (dígito verificador) antes de gastar qualquer chamada de API.
- Favoritos com etiquetas personalizadas (ex: "quente", "não atende", "já é cliente").
- Histórico das últimas consultas.
- "Colar e ir": ao colar uma lista de CNPJs, a ferramenta detecta e sugere verificar na hora.
- Atalhos de teclado: `Ctrl/Cmd+Enter` verifica, `Esc` fecha modais.

### Prospecção
- Busca de empresas ativas (e que ainda não são clientes Arcom) por cidade, UF e ramo (CNAE).
- Suporte a **múltiplas cidades e múltiplos ramos numa única busca**, com deduplicação automática de CNPJs repetidos.
- Ranking de concentração por bairro/cidade (alternativa leve a um mapa de calor, já que não há geocodificação configurada).
- Listas de prospecção salvas localmente, com nome, para recarregar depois sem repetir a busca.

### Busca por Razão Social
- Cole uma ou várias razões sociais e encontre os CNPJs correspondentes (via API da Casa dos Dados).

### Insight Comercial com IA
- Geração de insight comercial por empresa (resumo, pontos fortes, sinais de atenção, abordagem sugerida, perguntas de qualificação) usando **Groq** (modelo Llama 3.3 70B) combinado com busca na web via **Tavily**.
- Geração em lote para todas as empresas consultadas.
- **Ranking de leads por IA**: um único request classifica as melhores oportunidades entre as empresas já consultadas.
- Compartilhamento rápido do insight via WhatsApp ou e-mail.
- Chat livre com IA (usa o mesmo modelo), com histórico persistente no navegador.

### Exportação
- Excel/CSV (com seleção de quais colunas exportar), JSON e PDF.
- Abas separadas para dados de sócios e para os insights gerados por IA.

### Administração (multiusuário via Firebase)
- Login com e-mail/senha; novos usuários ficam **pendentes** até um admin aprovar.
- Painel de admin: aprovar/negar acessos, promover/remover administradores.
- Log das últimas consultas realizadas por todos os usuários.
- Indicador de "cota" (quantidade de CNPJs consultados no mês pelo usuário logado) — **apenas informativo**, não bloqueia consultas (ver seção de Segurança).

### Personalização
- Modo escuro/claro, cor de destaque customizável, partículas de fundo animadas.

---

## 🧱 Stack técnica

| Camada | Tecnologia |
|---|---|
| Front-end | HTML + CSS + JavaScript vanilla (sem build step) |
| Autenticação/Banco | Firebase Authentication + Firestore |
| Gráficos | Chart.js |
| Exportação | SheetJS (xlsx), html2pdf.js |
| IA | Groq (chat completions) + Tavily (busca na web) |
| Dados de CNPJ | BrasilAPI, OpenCNPJ, Casa dos Dados, API própria "Consulta CNPJ Arcom" |
| Hospedagem | GitHub Pages |

Não há back-end, bundler ou dependências de build — é um único arquivo `.html` que roda direto no navegador.

---

## 🚀 Como rodar localmente

1. Baixe o arquivo `.html` do projeto.
2. Configure as constantes no topo dos blocos `<script>` (ver seção abaixo).
3. Abra o arquivo direto no navegador, ou sirva com qualquer servidor estático:
   ```bash
   npx serve .
   # ou
   python3 -m http.server 8080
   ```

## ⚙️ Configuração necessária

No arquivo HTML, procure e preencha estas constantes antes de usar em produção:

| Constante | Onde usar | Serviço |
|---|---|---|
| `firebaseConfig` | Login e Firestore | [Console do Firebase](https://console.firebase.google.com) |
| `ADMIN_EMAIL` | E-mail do super admin | — |
| `ARCOM_API_KEY` | Consulta CNPJ Arcom | API interna da Arcom |
| `CASADOSDADOS_API_KEY` | Busca por razão social | [Casa dos Dados](https://casadosdados.com.br) |
| `GROQ_API_KEY` | Insight IA e chat | [console.groq.com/keys](https://console.groq.com/keys) |
| `TAVILY_API_KEY` | Busca na web para o insight | [tavily.com](https://tavily.com) |

Também é preciso configurar no **Console do Firebase**:
- Authentication → método de login por e-mail/senha habilitado.
- Firestore → coleções `usuarios` (perfis/aprovação) e `consultas_log` (log de uso) — criadas automaticamente pelo app no primeiro uso.

---

## 📄 Licença

Uso interno — Arcom Atacadista.
