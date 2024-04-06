package delivery

import "github.com/SanExpett/auto-catalog/pkg/models"

const (
	ResponseSuccessfulDeleteCar = "Автомобиль успешно удален"
)

type CarResponse struct {
	Status int         `json:"status"`
	Body   *models.Car `json:"body"`
}

func NewCarResponse(status int, body *models.Car) *CarResponse {
	return &CarResponse{
		Status: status,
		Body:   body,
	}
}

type CarListResponse struct {
	Status int           `json:"status"`
	Body   []*models.Car `json:"body"`
}

func NewCarListResponse(status int, body []*models.Car) *CarListResponse {
	return &CarListResponse{
		Status: status,
		Body:   body,
	}
}
