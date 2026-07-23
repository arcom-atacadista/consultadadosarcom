package enriquecimento

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/cnpj"
)

const maxLote = 500
const maxRefrescarPorListagem = 60 // trava de segurança (mesma do app antigo)

type Service struct {
	client *Trace360Client
	repo   *Repo
}

func NewService(client *Trace360Client, repo *Repo) *Service {
	return &Service{client: client, repo: repo}
}

// Enviar enfileira os CNPJs válidos na Trace360 e grava a posse do usuário.
func (s *Service) Enviar(ctx context.Context, uid string, cnpjsBrutos []string) ([]ItemEnviado, error) {
	limpos := make([]string, 0, len(cnpjsBrutos))
	for _, c := range cnpjsBrutos {
		l := cnpj.Limpar(c)
		if cnpj.Valido(l) {
			limpos = append(limpos, l)
		}
	}
	if len(limpos) > maxLote {
		limpos = limpos[:maxLote]
	}
	if len(limpos) == 0 {
		return nil, nil
	}

	resultados, err := s.client.EnviarLote(ctx, limpos)
	if err != nil {
		return nil, err
	}

	for _, item := range resultados {
		if item.ClienteID == "" {
			continue
		}
		e := &Enriquecimento{ID: uuid.NewString(), UID: uid, CNPJ: item.CNPJ, ClienteID: item.ClienteID, Status: "pendente"}
		if err := s.repo.Upsert(ctx, e); err != nil {
			slog.Error("falha ao gravar posse de enriquecimento", "erro", err)
		}
	}
	return resultados, nil
}

// Listar devolve os enriquecimentos do usuário, atualizando ao vivo o status
// dos que ainda estão em andamento (mesmo comportamento do app antigo).
func (s *Service) Listar(ctx context.Context, uid string) ([]Enriquecimento, error) {
	itens, err := s.repo.ListarPorUsuario(ctx, uid, maxRefrescarPorListagem)
	if err != nil {
		return nil, err
	}
	for i := range itens {
		if !StatusEmAndamento[itens[i].Status] {
			continue
		}
		st, err := s.client.Status(ctx, itens[i].ClienteID)
		if err != nil {
			continue // segue com o status que já tínhamos
		}
		if st.Status != "" && st.Status != itens[i].Status {
			itens[i].Status = st.Status
			itens[i].RazaoSocial = st.RazaoSocial
			if err := s.repo.AtualizarStatus(ctx, itens[i].ID, st.Status, st.RazaoSocial); err != nil {
				slog.Error("falha ao atualizar status do enriquecimento", "erro", err)
			}
		}
	}
	return itens, nil
}

// dono confere se o clienteId pertence ao uid — usado antes de qualquer
// endpoint de detalhe/dossiê/reprocessar.
func (s *Service) dono(ctx context.Context, uid, clienteID string) (*Enriquecimento, error) {
	e, err := s.repo.BuscarPorClienteID(ctx, clienteID)
	if err != nil {
		return nil, err
	}
	if e.UID != uid {
		return nil, ErrNaoEncontrado
	}
	return e, nil
}
