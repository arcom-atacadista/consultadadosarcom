package admin

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/atividades"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/auth"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/httputil"
)

// feedFetch/feedRetorno reproduzem exatamente os limites do app antigo:
// busca as 120 mais recentes, filtra por tipo em memória, mostra no máximo 60.
const (
	feedFetch   = 120
	feedRetorno = 60
)

type Handler struct {
	service    *Service
	atividades *atividades.Repo
}

func NewHandler(service *Service, atividadesRepo *atividades.Repo) *Handler {
	return &Handler{service: service, atividades: atividadesRepo}
}

// Routes monta as rotas admin-only (montadas sob /api/admin, já atrás de
// requireAuth+RequireAdmin no router).
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/dashboard", h.dashboard)
	r.Get("/atividades", h.atividadesFeed)
	r.Delete("/atividades/antigos", h.limparAntigos)
	return r
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	d, err := h.service.Dashboard(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao montar o dashboard")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, d)
}

func (h *Handler) atividadesFeed(w http.ResponseWriter, r *http.Request) {
	tipo := r.URL.Query().Get("tipo")
	ultimas, err := h.atividades.UltimasN(r.Context(), feedFetch)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao listar atividades")
		return
	}
	filtradas := ultimas
	if tipo != "" {
		filtradas = filtradas[:0]
		for _, a := range ultimas {
			if a.Tipo == tipo {
				filtradas = append(filtradas, a)
			}
		}
	}
	if len(filtradas) > feedRetorno {
		filtradas = filtradas[:feedRetorno]
	}
	httputil.WriteJSON(w, http.StatusOK, filtradas)
}

func (h *Handler) limparAntigos(w http.ResponseWriter, r *http.Request) {
	dias := 90
	if v := r.URL.Query().Get("dias"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			dias = n
		}
	}
	total, err := h.atividades.LimparAntigos(r.Context(), dias, atividades.TiposApagaveis)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao limpar atividades antigas")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"apagados": total})
}

// PresencaHandler expõe o heartbeat de presença — qualquer usuário aprovado
// chama, não é admin-only (a listagem de quem está online é que é admin-only,
// dentro do dashboard acima).
type PresencaHandler struct {
	service *PresencaService
}

func NewPresencaHandler(service *PresencaService) *PresencaHandler {
	return &PresencaHandler{service: service}
}

func (h *PresencaHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.FromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "faça login para continuar")
		return
	}
	if err := h.service.Heartbeat(r.Context(), claims.UID, claims.Nome, claims.Email); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao registrar presença")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *PresencaHandler) Remover(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.FromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "faça login para continuar")
		return
	}
	if err := h.service.Remover(r.Context(), claims.UID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao remover presença")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
