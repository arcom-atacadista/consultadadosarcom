package cnpj

// Socio é o mesmo formato que o app antigo já expõe no front — mantém
// paridade (docs/migracao/01-diagnostico-atual.md §3.1).
type Socio struct {
	NomeSocio            string `json:"nome_socio"`
	CPF                  string `json:"cpf"`
	QualificacaoSocio    string `json:"qualificacao_socio"`
	FaixaEtaria          string `json:"faixa_etaria"`
	DataEntradaSociedade string `json:"data_entrada_sociedade"`
}

// Empresa é o DTO normalizado que a API devolve pro front, único pras duas
// fontes (Arcom e BrasilAPI) — ver docs/migracao/03-contrato-api.md §3.
type Empresa struct {
	CNPJ                    string   `json:"cnpj"`
	Encontrado              bool     `json:"encontrado"`
	Erro                    string   `json:"erro,omitempty"`
	Situacao                string   `json:"situacao"`
	DataSituacaoCadastral   string   `json:"dataSituacaoCadastral"`
	MotivoSituacaoCadastral string   `json:"motivoSituacaoCadastral"`
	Razao                   string   `json:"razao"`
	NomeFantasia            string   `json:"nomeFantasia"`
	Porte                   string   `json:"porte"`
	ClienteArcom            string   `json:"clienteArcom"` // "Sim" | "Não" | "—"
	Natureza                string   `json:"natureza"`
	Atividade               string   `json:"atividade"`
	MatrizFilial            string   `json:"matrizFilial"`
	UF                      string   `json:"uf"`
	Municipio               string   `json:"municipio"`
	DataInicio              string   `json:"dataInicio"`
	CapitalSocial           string   `json:"capitalSocial"`
	Endereco                string   `json:"endereco"`
	CEP                     string   `json:"cep"`
	Telefone                string   `json:"telefone"`
	Email                   string   `json:"email"`
	Socios                  []Socio  `json:"socios"`
	Simples                 string   `json:"simples"`
	MEI                     string   `json:"mei"`
	Latitude                *float64 `json:"latitude"`
	Longitude               *float64 `json:"longitude"`
	API                     string   `json:"api"` // "Consulta CNPJ Arcom" | "Brasil API"
}
