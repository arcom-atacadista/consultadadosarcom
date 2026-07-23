package prospeccao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/atividades"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/auth"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/httputil"
)

var validate = validator.New()

type Handler struct {
	service    *Service
	repo       *Repo
	buscador   *Buscador
	atividades *atividades.Service
}

func NewHandler(service *Service, repo *Repo, buscador *Buscador, atividadesService *atividades.Service) *Handler {
	return &Handler{service: service, repo: repo, buscador: buscador, atividades: atividadesService}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/municipios", h.municipios)
	r.Post("/buscar", h.buscar)
	r.Get("/listas", h.listarListas)
	r.Post("/listas", h.criarLista)
	r.Delete("/listas/{id}", h.deletarLista)
	return r
}

func (h *Handler) municipios(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		httputil.WriteError(w, http.StatusBadRequest, "informe a cidade em ?q=")
		return
	}
	uf := r.URL.Query().Get("uf")

	municipio, err := ResolverMunicipio(r.Context(), h.buscador.arcom, q, uf)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, municipio)
}

type buscarInput struct {
	Cidades []CidadeFiltro `json:"cidades" validate:"required,min=1,max=20,dive"`
	CNAEs   []string       `json:"cnaes" validate:"required,min=1,max=15"`
}

func (h *Handler) buscar(w http.ResponseWriter, r *http.Request) {
	var in buscarInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := validate.Struct(in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "informe ao menos 1 cidade (máx. 20) e 1 ramo (máx. 15)")
		return
	}
	for _, c := range in.Cidades {
		if c.Cidade == "" || c.UF == "" {
			httputil.WriteError(w, http.StatusBadRequest, "cada cidade precisa de UF (formato Cidade,UF)")
			return
		}
	}

	uid := uidDoContexto(r.Context())
	prospects, err := h.service.Buscar(r.Context(), uid, BuscarInput{Cidades: in.Cidades, CNAEs: in.CNAEs})
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, err.Error())
		return
	}
	if claims, ok := auth.FromContext(r.Context()); ok {
		cidadesTxt := make([]string, len(in.Cidades))
		for i, c := range in.Cidades {
			cidadesTxt[i] = c.Cidade + "/" + c.UF
		}
		detalhe := fmt.Sprintf("%d prospect(s) · %s · %d ramo(s)", len(prospects), strings.Join(cidadesTxt, ", "), len(in.CNAEs))
		h.atividades.Registrar(r.Context(), claims.UID, claims.Nome, claims.Email, atividades.TipoProspeccao, detalhe, nil)
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"prospects": prospects})
}

type criarListaInput struct {
	Nome     string `json:"nome" validate:"required"`
	Assessor string `json:"assessor" validate:"required"`
	Filtros  any    `json:"filtros" validate:"required"`
	Itens    any    `json:"itens" validate:"required"`
}

func (h *Handler) criarLista(w http.ResponseWriter, r *http.Request) {
	var in criarListaInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := validate.Struct(in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "informe nome, assessor, filtros e itens")
		return
	}
	claims, _ := auth.FromContext(r.Context())
	lista := ListaProspeccao{
		UID: claims.UID, Nome: in.Nome, Assessor: in.Assessor,
		NomeUsuario: claims.Nome, Email: claims.Email,
		Filtros: in.Filtros, Itens: in.Itens,
	}
	if err := h.repo.CriarLista(r.Context(), &lista); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao salvar lista")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, lista)
}

func (h *Handler) listarListas(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.FromContext(r.Context())
	listas, err := h.repo.ListarListas(r.Context(), claims.UID, claims.IsAdmin)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao listar")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, listas)
}

func (h *Handler) deletarLista(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims, _ := auth.FromContext(r.Context())

	lista, err := h.repo.BuscarLista(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "lista não encontrada")
		return
	}
	if lista.UID != claims.UID && !claims.IsAdmin {
		httputil.WriteError(w, http.StatusForbidden, "só o dono ou um admin pode apagar essa lista")
		return
	}
	if err := h.repo.DeletarLista(r.Context(), id); err != nil {
		if errors.Is(err, ErrNaoEncontrado) {
			httputil.WriteError(w, http.StatusNotFound, "lista não encontrada")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao apagar lista")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func uidDoContexto(ctx context.Context) string {
	if claims, ok := auth.FromContext(ctx); ok {
		return claims.UID
	}
	return ""
}
