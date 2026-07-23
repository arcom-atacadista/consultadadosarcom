package prospeccao

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/atividades"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/auth"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/httputil"
)

type PreCadastroHandler struct {
	repo       *Repo
	atividades *atividades.Service
}

func NewPreCadastroHandler(repo *Repo, atividadesService *atividades.Service) *PreCadastroHandler {
	return &PreCadastroHandler{repo: repo, atividades: atividadesService}
}

func (h *PreCadastroHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.listar)
	r.Post("/", h.criar)
	r.Patch("/{id}", h.atualizar)
	r.Delete("/{id}", h.deletar)
	return r
}

type criarPreCadastroInput struct {
	CNPJ     string `json:"cnpj" validate:"required"`
	Razao    string `json:"razao"`
	Endereco string `json:"endereco"`
	Contato  string `json:"contato"`
	Notas    string `json:"notas"`
}

func (h *PreCadastroHandler) criar(w http.ResponseWriter, r *http.Request) {
	var in criarPreCadastroInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := validate.Struct(in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "informe ao menos o CNPJ")
		return
	}
	uid := uidDoContexto(r.Context())
	p := PreCadastro{
		CNPJ: in.CNPJ, Razao: in.Razao, Endereco: in.Endereco,
		Contato: in.Contato, Notas: in.Notas, Status: "novo", AutorUID: uid,
	}
	if err := h.repo.CriarPreCadastro(r.Context(), &p); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao criar pré-cadastro")
		return
	}
	if claims, ok := auth.FromContext(r.Context()); ok {
		nome := p.Razao
		if nome == "" {
			nome = in.CNPJ
		}
		h.atividades.Registrar(r.Context(), claims.UID, claims.Nome, claims.Email, atividades.TipoPreCadastro,
			fmt.Sprintf("Pré-cadastrou %s", nome), nil)
	}
	httputil.WriteJSON(w, http.StatusCreated, p)
}

func (h *PreCadastroHandler) listar(w http.ResponseWriter, r *http.Request) {
	out, err := h.repo.ListarPreCadastros(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao listar pré-cadastros")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

type atualizarPreCadastroInput struct {
	Status string `json:"status" validate:"omitempty,oneof=novo em_contato convertido descartado"`
	Notas  string `json:"notas"`
}

func (h *PreCadastroHandler) atualizar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var in atualizarPreCadastroInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := validate.Struct(in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "status inválido")
		return
	}
	if err := h.repo.AtualizarPreCadastro(r.Context(), id, in.Status, in.Notas); err != nil {
		if errors.Is(err, ErrNaoEncontrado) {
			httputil.WriteError(w, http.StatusNotFound, "pré-cadastro não encontrado")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao atualizar")
		return
	}
	if claims, ok := auth.FromContext(r.Context()); ok {
		h.atividades.Registrar(r.Context(), claims.UID, claims.Nome, claims.Email, atividades.TipoPreCadastro,
			fmt.Sprintf("Atualizou pré-cadastro %s", id), nil)
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *PreCadastroHandler) deletar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.DeletarPreCadastro(r.Context(), id); err != nil {
		if errors.Is(err, ErrNaoEncontrado) {
			httputil.WriteError(w, http.StatusNotFound, "pré-cadastro não encontrado")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao remover")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
