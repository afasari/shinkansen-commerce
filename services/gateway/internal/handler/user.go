package handler

import (
	"context"
	"encoding/json"
	"net/http"

	userpb "github.com/afasari/shinkansen-commerce/gen/proto/go/user"
	"google.golang.org/grpc"
)

type UserHandler struct {
	client userpb.UserServiceClient
}

func NewUserHandler(conn *grpc.ClientConn) *UserHandler {
	return &UserHandler{
		client: userpb.NewUserServiceClient(conn),
	}
}

func (h *UserHandler) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/users/register", h.handleRegister)
	mux.HandleFunc("/v1/users/login", h.handleLogin)
	mux.HandleFunc("/v1/users/", h.handleUser)
	mux.HandleFunc("/v1/users/me/addresses", h.handleMyAddresses)
	mux.HandleFunc("/v1/addresses/", h.handleAddress)
}

func (h *UserHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req userpb.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.RegisterUser(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *UserHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req userpb.LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.LoginUser(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) handleUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	path := r.URL.Path[len("/v1/users/"):]

	if path == "me" {
		switch r.Method {
		case http.MethodGet:
			h.getCurrentUser(w, r, ctx)
		case http.MethodPut:
			h.updateCurrentUser(w, r, ctx)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getUser(w, r, ctx, path)
	case http.MethodPut:
		h.updateUser(w, r, ctx, path)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) handleMyAddresses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listAddresses(w, r, ctx, userID.(string))
	case http.MethodPost:
		h.addAddress(w, r, ctx, userID.(string))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) handleAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	addressID := r.URL.Path[len("/v1/addresses/"):]
	if addressID == "" {
		http.Error(w, "Address ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		h.updateAddress(w, r, ctx, addressID)
	case http.MethodDelete:
		h.deleteAddress(w, r, ctx, addressID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) getCurrentUser(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req := &userpb.GetUserRequest{UserId: userID.(string)}
	resp, err := h.client.GetUser(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request, ctx context.Context, userID string) {
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	req := &userpb.GetUserRequest{UserId: userID}
	resp, err := h.client.GetUser(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) updateUser(w http.ResponseWriter, r *http.Request, ctx context.Context, userID string) {
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	var req userpb.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.UserId = userID

	resp, err := h.client.UpdateUser(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) updateCurrentUser(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req userpb.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.UserId = userID.(string)

	resp, err := h.client.UpdateUser(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) listAddresses(w http.ResponseWriter, r *http.Request, ctx context.Context, userID string) {
	req := &userpb.ListAddressesRequest{UserId: userID}
	resp, err := h.client.ListAddresses(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) addAddress(w http.ResponseWriter, r *http.Request, ctx context.Context, userID string) {
	var req userpb.AddAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.UserId = userID

	resp, err := h.client.AddAddress(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *UserHandler) updateAddress(w http.ResponseWriter, r *http.Request, ctx context.Context, addressID string) {
	var req userpb.UpdateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.AddressId = addressID

	resp, err := h.client.UpdateAddress(ctx, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) deleteAddress(w http.ResponseWriter, r *http.Request, ctx context.Context, addressID string) {
	req := &userpb.DeleteAddressRequest{AddressId: addressID}
	_, err := h.client.DeleteAddress(ctx, req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
