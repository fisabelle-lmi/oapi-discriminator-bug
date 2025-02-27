package server

import (
	"context"
	"github.com/fisabelle-lmi/oapi-discriminator-bug/pkg/api/common"
	privateApi "github.com/fisabelle-lmi/oapi-discriminator-bug/pkg/api/private"
)

type apiServer struct{}

func (s *apiServer) CreatePet(_ context.Context, _ privateApi.CreatePetRequestObject) (privateApi.CreatePetResponseObject, error) {
	return privateApi.CreatePet406JSONResponse{
		N406: common.N406{
			Message:   "Not implemented",
			ErrorCode: common.ErrorCodeNOTIMPLEMENTED,
		},
	}, nil
}
func (s *apiServer) GetPets(_ context.Context, _ privateApi.GetPetsRequestObject) (privateApi.GetPetsResponseObject, error) {
	return privateApi.GetPets406JSONResponse{
		N406: common.N406{
			Message:   "Not implemented",
			ErrorCode: common.ErrorCodeNOTIMPLEMENTED,
		},
	}, nil
}

func NewPrivateApiServer() privateApi.StrictServerInterface {
	return &apiServer{}
}
