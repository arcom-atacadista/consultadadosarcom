package atividades

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/httputil"
)

// ClaimsFn lê o usuário autenticado do contexto sem o pacote atividades
// precisar importar auth (evitaria ciclo: auth já importa atividades pra
// logar conta_criada/login).
type ClaimsFn func(ctx context.Context) (uid, nome, email string, ok bool)

// Handler expõe só o registro de "gerou PDF de indicação" — o único tipo de
// atividade do app antigo que nasce inteiramente no navegador (o PDF é
// window.print() client-side, sem round-trip natural pro backend).
type Handler struct {
	service *Service
	claims  ClaimsFn
}

func NewHandler(service *Service, claimsFn ClaimsFn) *Handler {
	return &Handler{service: service, claims: claimsFn}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/pdf-indicacao", h.pdfIndicacao)
	return r
}

type pdfIndicacaoInput struct {
	CNPJs []string `json:"cnpjs" validate:"required,min=1"`
}

func (h *Handler) pdfIndicacao(w http.ResponseWriter, r *http.Request) {
	var in pdfIndicacaoInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if len(in.CNPJs) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "informe ao menos 1 CNPJ")
		return
	}
	uid, nome, email, ok := h.claims(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "faça login para continuar")
		return
	}
	h.service.Registrar(r.Context(), uid, nome, email, TipoPDFIndicacao,
		fmt.Sprintf("Gerou PDF de indicação com %d empresa(s)", len(in.CNPJs)),
		map[string]any{"cnpjs": in.CNPJs})
	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
