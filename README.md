# CDA — Consulta Dados Arcom

Aplicação web (single-page, HTML/CSS/JS puro) para consulta, prospecção e enriquecimento comercial de empresas via CNPJ. Feita para uso interno da Arcom Atacadista.

🔗 Deploy: https://arcom-atacadista.github.io/consultadadosarcom/

---

## ✨ Funcionalidades

### Consulta de CNPJ
- Consulta individual ou em lote (cole vários CNPJs, um por linha).
- Fonte de dados: **Consulta CNPJ Arcom** (API própria, com autenticação JWT e suporte a lote de até 1000 CNPJs por chamada).
- Mostra na hora se a empresa **já é cliente Arcom**.
- Exibe todos os campos que a API retorna (razão social, nome fantasia, situação, sócios, endereço, telefone, etc.).
- Cache local de 24h: reconsultar o mesmo CNPJ não gasta uma nova chamada.
- Validação de CNPJ no cliente (dígito verificador) antes de gastar qualquer chamada de API.
- Favoritos com etiquetas personalizadas e histórico das últimas consultas.
- Atalhos de teclado: `Ctrl/Cmd+Enter` verifica, `Esc` fecha modais.

### Prospecção
- Busca de empresas ativas (e que ainda não são clientes Arcom) por cidade, UF e ramo (CNAE).
- Suporte a **múltiplas cidades e múltiplos ramos numa única busca**, com deduplicação automática.
- **Corte de grandes redes**: filtra automaticamente as grandes cadeias (pelo número da filial), focando em médio e pequeno porte.
- Ranking de concentração por bairro/cidade.
- Botões de contato clicáveis por linha (ligar, WhatsApp, e-mail) e pré-cadastro direto.
- Listas de prospecção salvas localmente, para recarregar depois sem repetir a busca.

### Provável loja física (validação visual)
- Cada prospect recebe um **score de 0 a 100** ("provável loja física") com semáforo 🟢/🟡/🔴, calculado a partir de sinais como ramo (CNAE), nome fantasia, endereço com número, telefone fixo e tempo de atividade.
- A lista é ordenada pelos leads de maior potencial.
- Ao abrir o mapa de uma empresa, o modal já mostra a **fachada (Street View)** embutida — o jeito definitivo de confirmar se é loja física de verdade, sem sair do site. Abas **Fachada** / **Mapa** e faixa de veredicto com os sinais que pesaram.
- Tudo aberto **pelo endereço** do CNPJ (as coordenadas só ordenam por proximidade nos bastidores).

### Pré-cadastro de clientes
- Cadastro auto-preenchido com os dados da consulta (razão, CNPJ, endereço, contato).
- Lista de acompanhamento com busca, notas, edição e status, dentro das abas de Consulta e Prospecção.

### Indicação de clientes em PDF
- Exporta um PDF de indicação agrupado por ramo, com CNPJ, endereço, CEP, sócios e **telefone**.
- Seleção dos prospects que entram no PDF.
- Ordenação por proximidade a partir do ponto de partida do assessor (bairro ou GPS).
- Rota de visita no **Google Maps/Waze** (link clicável, QR code e botões no PDF).

### Insight Comercial com IA
- Geração de insight comercial por empresa usando **Groq** (Llama 3.3 70B) combinado com busca na web via **Tavily**.
- Geração em lote, **ranking de leads por IA** e chat livre com histórico no navegador.

### Faturamento de postos (contas a pagar)
- Aba **Faturamento** para o controle do faturamento que os postos enviam (hoje por e-mail/WhatsApp).
- **Upload dos anexos** (boleto, relatório e nota) — a **IA lê** o boleto/nota (via **Groq** — texto do PDF ou visão para foto/scan) e preenche **posto, valor, vencimento, banco e nº da nota** pra você conferir. Conferência offline pela **linha digitável** do boleto (banco/valor/vencimento) como rede de segurança.
- **Conferência de documentos**: marca se veio **boleto + relatório + nota** ou o que faltou.
- **Conciliação de quantidade** (litros do relatório × conferido no **Web SIA**): só libera a **programação de pagamento** quando bate — respeitando o mínimo de **2 dias úteis**.
- **Dashboard**: todas as faturas **por data de pagamento**, pago/não pago, totais e gráficos (por dia e por posto).
- Aba de **faturas pendentes por documento** (boleto/relatório/nota faltando).
- **Assistente em linguagem natural** (Groq): perguntas como *"quanto vence essa semana do posto X?"* ou *"o que está faltando de documento?"*.
- Observação: nesta versão (100% front-end) os **arquivos não são arquivados** — fica só o registro conferido. Captura automática do e-mail/WhatsApp e arquivamento dos anexos ficam para uma fase 2 (com back-end).

### Exportação
- Excel/CSV (com seleção de colunas), JSON e PDF, com abas separadas para sócios e insights.

### Administração (multiusuário via Firebase)
- Login com e-mail/senha; novos usuários ficam **pendentes** até um admin aprovar.
- Aprovar/negar acessos e promover/remover administradores.

### Painel do administrador (dashboard)
- Aba exclusiva de admin que acompanha tudo no site: **contas criadas, contas pendentes, logins, consultas, prospecções e PDFs gerados**.
- **Histórico de atividades** completo (com filtro por tipo) e nome de quem fez cada ação.
- Tabela de **atividade por usuário** (logins/consultas/prospecções/PDFs por pessoa).

### Interface
- Tela inicial (cover) apresentando a ferramenta, tela de login e logo oficial da Arcom.
- Tema escuro por padrão (claro opcional), com cor de destaque customizável e partículas de fundo.
- Barra de navegação **recolhível** (vira uma mini-barra só de ícones) para expandir a área de trabalho.
- Micro-animações suaves (respeitando `prefers-reduced-motion`).

---

## 🧱 Stack técnica

| Camada | Tecnologia |
|---|---|
| Front-end | HTML + CSS + JavaScript vanilla (sem build step) |
| Autenticação/Banco | Firebase Authentication + Firestore |
| Gráficos | Chart.js |
| Exportação | SheetJS (xlsx), html2pdf.js, QR Code (qrcodejs) |
| Mapa/validação | Google Maps + Street View (embed, por endereço) |
| IA | Groq (chat completions) + Tavily (busca na web) |
| Dados de CNPJ | API própria "Consulta CNPJ Arcom", BrasilAPI |
| Hospedagem | GitHub Pages |

Não há back-end, bundler ou dependências de build — é um único arquivo `index.html` que roda direto no navegador.

---

## 🚀 Como rodar localmente

1. Baixe o arquivo `index.html` do projeto.
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
| `GROQ_API_KEY` | Insight IA e chat | [console.groq.com/keys](https://console.groq.com/keys) |
| `TAVILY_API_KEY` | Busca na web para o insight | [tavily.com](https://tavily.com) |

Também é preciso configurar no **Console do Firebase**:
- Authentication → método de login por e-mail/senha habilitado.
- Firestore → coleções usadas pelo app (criadas automaticamente no primeiro uso):
  - `usuarios` — perfis, status de aprovação e flag de admin
  - `preCadastros` — pré-cadastros de clientes
  - `consultas_log` — log de consultas (usado no contador de uso mensal)
  - `atividades_log` — histórico geral de atividades (login, consulta, prospecção, PDF, etc.)
  - `faturas` — faturamento dos postos (contas a pagar): posto, valor, vencimento, documentos, conciliação e status de pagamento (cada usuário vê só as próprias; admin vê todas)

### Regras do Firestore (exemplo)

```
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {

    function isAdmin() {
      return request.auth != null && (
        request.auth.token.email == 'SEU_SUPER_ADMIN@exemplo.com' ||
        get(/databases/$(database)/documents/usuarios/$(request.auth.uid)).data.isAdmin == true
      );
    }

    match /usuarios/{userId} {
      allow create: if request.auth.uid == userId
                     && request.resource.data.status == 'pendente';
      allow read: if request.auth.uid == userId || isAdmin();
      allow update: if isAdmin();
    }

    match /preCadastros/{docId} {
      allow read, write: if request.auth != null
        && get(/databases/$(database)/documents/usuarios/$(request.auth.uid)).data.status == 'aprovado';
    }

    match /consultas_log/{docId} {
      allow create: if request.auth != null;
      allow read: if request.auth != null;
    }

    match /atividades_log/{docId} {
      allow create: if request.auth != null;
      allow read: if request.auth != null;
    }
  }
}
```

---

## 📄 Licença

Uso interno — Arcom Atacadista.
