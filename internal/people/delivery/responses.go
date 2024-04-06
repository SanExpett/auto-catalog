package delivery

import "github.com/SanExpett/auto-catalog/pkg/models"

const (
	ResponseSuccessfulDeletePeople = "Человек успешно удален"
)

type PeopleResponse struct {
	Status int            `json:"status"`
	Body   *models.People `json:"body"`
}

func NewPeopleResponse(status int, body *models.People) *PeopleResponse {
	return &PeopleResponse{
		Status: status,
		Body:   body,
	}
}
