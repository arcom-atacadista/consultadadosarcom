package enriquecimento

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/auth"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/httputil"
)

var validate = validator.New()

type Handler struct {
	service *Service
	client  *Trace360Client
}

func NewHandler(service *Service, client *Trace360Client) *Handler {
	return &Handler{service: service, client: client}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.enviar)
	r.Get("/", h.listar)
	r.Get("/{clienteId}", h.detalhe)
	r.Get("/{clienteId}/progresso", h.progresso)
	r.Get("/{clienteId}/dossie", h.dossie)
	r.Get("/{clienteId}/resultado", h.resultado)
	r.Post("/{clienteId}/reprocessar", h.reprocessar)
	return r
}

type enviarInput struct {
	CNPJs []string `json:"cnpjs" validate:"required,min=1,max=500"`
}

func (h *Handler) enviar(w http.ResponseWriter, r *http.Request) {
	var in enviarInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := validate.Struct(in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "informe ao menos 1 CNPJ (máx. 500)")
		return
	}
	uid := uidDoContexto(r.Context())
	resultados, err := h.service.Enviar(r.Context(), uid, in.CNPJs)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "falha ao enviar pra Trace360: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"resultados": resultados})
}

func (h *Handler) listar(w http.ResponseWriter, r *http.Request) {
	uid := uidDoContexto(r.Context())
	itens, err := h.service.Listar(r.Context(), uid)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao listar enriquecimentos")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, itens)
}

func (h *Handler) buscarDono(w http.ResponseWriter, r *http.Request) (*Enriquecimento, bool) {
	uid := uidDoContexto(r.Context())
	clienteID := chi.URLParam(r, "clienteId")
	e, err := h.service.dono(r.Context(), uid, clienteID)
	if err != nil {
		if errors.Is(err, ErrNaoEncontrado) {
			httputil.WriteError(w, http.StatusNotFound, "enriquecimento não encontrado")
		} else {
			httputil.WriteError(w, http.StatusInternalServerError, "falha ao buscar enriquecimento")
		}
		return nil, false
	}
	return e, true
}

func (h *Handler) detalhe(w http.ResponseWriter, r *http.Request) {
	e, ok := h.buscarDono(w, r)
	if !ok {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, e)
}

func (h *Handler) progresso(w http.ResponseWriter, r *http.Request) {
	e, ok := h.buscarDono(w, r)
	if !ok {
		return
	}
	p, err := h.client.Progresso(r.Context(), e.ClienteID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "falha ao buscar o fluxo: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, p)
}

func (h *Handler) dossie(w http.ResponseWriter, r *http.Request) {
	e, ok := h.buscarDono(w, r)
	if !ok {
		return
	}
	bytes, contentType, err := h.client.Dossie(r.Context(), e.ClienteID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "falha ao baixar o dossiê: "+err.Error())
		return
	}
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Content-Disposition", "attachment; filename=dossie-"+e.ClienteID+".pdf")
	w.Write(bytes)
}

func (h *Handler) resultado(w http.ResponseWriter, r *http.Request) {
	e, ok := h.buscarDono(w, r)
	if !ok {
		return
	}
	raw, err := h.client.Resultado(r.Context(), e.ClienteID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "falha ao buscar o resultado: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func (h *Handler) reprocessar(w http.ResponseWriter, r *http.Request) {
	e, ok := h.buscarDono(w, r)
	if !ok {
		return
	}
	if err := h.client.Reprocessar(r.Context(), e.ClienteID); err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "falha ao reprocessar: "+err.Error())
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
