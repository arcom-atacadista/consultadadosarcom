package prospeccao

// Prospect é o DTO normalizado que a busca de prospecção devolve — já com o
// score de loja física calculado no backend (docs/migracao/01 §6, regra de
// negócio que não pode ficar exposta no navegador).
type Prospect struct {
	CNPJ         string   `json:"cnpj"`
	Razao        string   `json:"razao"`
	NomeFantasia string   `json:"nomeFantasia"`
	Atividade    string   `json:"atividade"`
	CNAECodigo   string   `json:"cnaeCodigo"`
	Bairro       string   `json:"bairro"`
	Endereco     string   `json:"endereco"`
	CEP          string   `json:"cep"`
	Telefone     string   `json:"telefone"`
	Email        string   `json:"email"`
	Porte        string   `json:"porte"`
	DataInicio   string   `json:"dataInicio"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	Cidade       string   `json:"cidade"`
	UF           string   `json:"uf"`

	// Sinais calculados no backend — o front só exibe, não recalcula.
	Score       int      `json:"score"`       // 0-100
	Temperatura string   `json:"temperatura"` // "quente" | "morno" | "frio"
	Sinais      []string `json:"sinais"`      // chips explicando o score
	RedeGrande  bool     `json:"redeGrande"`  // grande rede (corte de redes)
	Nova        bool     `json:"nova"`        // aberta nos últimos 12 meses
	MesesAtivo  *int     `json:"mesesAtivo,omitempty"`
}

type CidadeFiltro struct {
	Cidade string `json:"cidade"`
	UF     string `json:"uf"`
}
