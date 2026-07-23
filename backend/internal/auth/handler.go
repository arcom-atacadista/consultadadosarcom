package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/httputil"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/usuarios"
)

var validate = validator.New()

type registerInput struct {
	Email string `json:"email" validate:"required,email"`
	Senha string `json:"senha" validate:"required,min=8"`
	Nome  string `json:"nome" validate:"required"`
}

type loginInput struct {
	Email string `json:"email" validate:"required,email"`
	Senha string `json:"senha" validate:"required"`
}

type senhaInput struct {
	Atual string `json:"atual" validate:"required"`
	Nova  string `json:"nova" validate:"required,min=8"`
}

type Handler struct {
	service *Service
	repo    *usuarios.Repo
}

func NewHandler(service *Service, repo *usuarios.Repo) *Handler {
	return &Handler{service: service, repo: repo}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in registerInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := validate.Struct(in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "e-mail, senha (mín. 8) e nome são obrigatórios")
		return
	}
	u, err := h.service.Registrar(r.Context(), in.Email, in.Senha, in.Nome)
	if err != nil {
		if errors.Is(err, ErrEmailEmUso) {
			httputil.WriteError(w, http.StatusConflict, "e-mail já cadastrado")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao criar conta")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, u)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in loginInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := validate.Struct(in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "informe e-mail e senha")
		return
	}
	token, u, err := h.service.Login(r.Context(), in.Email, in.Senha)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "e-mail ou senha inválidos")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"token": token, "usuario": u})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, _ := FromContext(r.Context())
	u, err := h.repo.BuscarPorID(r.Context(), claims.UID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "usuário não encontrado")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"usuario":  u,
		"isAdmin":  claims.IsAdmin,
		"aprovado": claims.Aprovado,
	})
}

func (h *Handler) TrocarSenha(w http.ResponseWriter, r *http.Request) {
	claims, _ := FromContext(r.Context())
	var in senhaInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := validate.Struct(in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "informe a senha atual e a nova (mín. 8)")
		return
	}
	if err := h.service.TrocarSenha(r.Context(), claims.UID, in.Atual, in.Nova); err != nil {
		if errors.Is(err, ErrCredenciaisInvalidas) {
			httputil.WriteError(w, http.StatusUnauthorized, "senha atual incorreta")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao trocar a senha")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
