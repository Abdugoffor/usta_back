package telegram_auth_dto

import user_model "main_service/module/user_service/model"

type StartResponse struct {
	Token    string `json:"token"`
	DeepLink string `json:"deep_link"`
}

type StatusResponse struct {
	Status string           `json:"status"`
	JWT    string           `json:"jwt,omitempty"`
	User   *user_model.User `json:"user,omitempty"`
}
