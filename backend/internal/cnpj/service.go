package cnpj

import (
	"context"
	"log/slog"
)

type Service struct {
	arcom     *ArcomClient
	brasilAPI *BrasilAPIClient
	cache     *Cache
	repo      *Repo
}

func NewService(arcom *ArcomClient, brasilAPI *BrasilAPIClient, cache *Cache, repo *Repo) *Service {
	return &Service{arcom: arcom, brasilAPI: brasilAPI, cache: cache, repo: repo}
}

// Consultar valida, deduplica, resolve pelo cache o que der e só chama a
// fonte externa pro que sobrou — na ordem dos CNPJs recebidos.
func (s *Service) Consultar(ctx context.Context, uid string, cnpjsBrutos []string, fonte string) ([]Empresa, error) {
	if fonte != "arcom" && fonte != "brasilapi" {
		fonte = "arcom"
	}

	ordem := make([]string, 0, len(cnpjsBrutos))
	vistos := map[string]bool{}
	for _, raw := range cnpjsBrutos {
		limpo := Limpar(raw)
		if limpo == "" || vistos[limpo] {
			continue
		}
		vistos[limpo] = true
		ordem = append(ordem, limpo)
	}

	resultado := make(map[string]Empresa, len(ordem))
	var buscar []string

	for _, cnpj := range ordem {
		if !Valido(cnpj) {
			resultado[cnpj] = Empresa{CNPJ: cnpj, Encontrado: false, Erro: "CNPJ inválido"}
			continue
		}
		if cached, ok := s.cache.Buscar(ctx, fonte, cnpj); ok {
			resultado[cnpj] = *cached
			continue
		}
		buscar = append(buscar, cnpj)
	}

	if len(buscar) > 0 {
		var frescos []Empresa
		var err error
		if fonte == "arcom" {
			frescos, err = s.consultarArcom(ctx, buscar)
		} else {
			frescos, err = s.consultarBrasilAPI(ctx, buscar)
		}
		if err != nil {
			return nil, err
		}

		buscados := make([]string, 0, len(frescos))
		for _, e := range frescos {
			resultado[e.CNPJ] = e
			if e.Encontrado {
				s.cache.Salvar(ctx, fonte, e)
			}
			buscados = append(buscados, e.CNPJ)
		}

		if uid != "" {
			if err := s.repo.RegistrarConsultas(ctx, uid, buscados, fonte); err != nil {
				slog.Error("falha ao registrar consultas_log", "erro", err)
			}
		}
	}

	lista := make([]Empresa, 0, len(ordem))
	for _, cnpj := range ordem {
		lista = append(lista, resultado[cnpj])
	}
	return lista, nil
}

func (s *Service) consultarArcom(ctx context.Context, cnpjs []string) ([]Empresa, error) {
	itens, err := s.arcom.ConsultarLote(ctx, cnpjs)
	if err != nil {
		return nil, err
	}
	out := make([]Empresa, 0, len(itens))
	for _, item := range itens {
		out = append(out, mapArcomItem(item))
	}
	return out, nil
}

func (s *Service) consultarBrasilAPI(ctx context.Context, cnpjs []string) ([]Empresa, error) {
	out := make([]Empresa, 0, len(cnpjs))
	for _, cnpj := range cnpjs {
		e, err := s.brasilAPI.Consultar(ctx, cnpj)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}
