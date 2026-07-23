package cnpj

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/auth"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/httputil"
)

var validate = validator.New()

type consultarInput struct {
	CNPJs []string `json:"cnpjs" validate:"required,min=1,max=1000,dive,required"`
	Fonte string   `json:"fonte" validate:"omitempty,oneof=arcom brasilapi"`
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Consultar(w http.ResponseWriter, r *http.Request) {
	var in consultarInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := validate.Struct(in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "informe ao menos 1 CNPJ (máx. 1000)")
		return
	}
	fonte := in.Fonte
	if fonte == "" {
		fonte = "arcom"
	}

	uid := ""
	if claims, ok := auth.FromContext(r.Context()); ok {
		uid = claims.UID
	}

	resultado, err := h.service.Consultar(r.Context(), uid, in.CNPJs, fonte)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "falha ao consultar CNPJ: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"resultados": resultado})
}
