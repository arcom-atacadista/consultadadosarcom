package usuarios

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/httputil"
)

var validate = validator.New()

type patchInput struct {
	Status  *string `json:"status" validate:"omitempty,oneof=pendente aprovado negado"`
	IsAdmin *bool   `json:"isAdmin"`
}

// Handler expõe as rotas de administração de contas (GET/PATCH/DELETE),
// montadas sob /api/usuarios só para quem já passou pelo middleware de admin.
type Handler struct {
	repo *Repo
}

func NewHandler(repo *Repo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.listar)
	r.Patch("/{id}", h.atualizar)
	r.Delete("/{id}", h.deletar)
	return r
}

func (h *Handler) listar(w http.ResponseWriter, r *http.Request) {
	us, err := h.repo.Listar(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao listar usuários")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, us)
}

func (h *Handler) atualizar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var in patchInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := validate.Struct(in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "status deve ser 'pendente', 'aprovado' ou 'negado'")
		return
	}
	if err := h.repo.Atualizar(r.Context(), id, in.Status, in.IsAdmin); err != nil {
		if errors.Is(err, ErrNaoEncontrado) {
			httputil.WriteError(w, http.StatusNotFound, "usuário não encontrado")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao atualizar usuário")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) deletar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Deletar(r.Context(), id); err != nil {
		if errors.Is(err, ErrNaoEncontrado) {
			httputil.WriteError(w, http.StatusNotFound, "usuário não encontrado")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao remover usuário")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
