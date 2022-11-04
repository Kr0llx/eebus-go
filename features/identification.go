package features

import (
	"fmt"

	"github.com/DerAndereAndi/eebus-go/service"
	"github.com/DerAndereAndi/eebus-go/spine"
	"github.com/DerAndereAndi/eebus-go/spine/model"
)

type IdentificationType struct {
	Identifier string
	Type       model.IdentificationTypeType
}

type Identification struct {
	*FeatureImpl
}

func NewIdentification(service *service.EEBUSService, entity *spine.EntityRemoteImpl) (*Identification, error) {
	feature, err := NewFeatureImpl(model.FeatureTypeTypeIdentification, service, entity)
	if err != nil {
		return nil, err
	}

	i := &Identification{
		FeatureImpl: feature,
	}

	return i, nil
}

// request FunctionTypeIdentificationListData from a remote entity
func (i *Identification) Request() (*model.MsgCounterType, error) {
	msgCounter, err := i.requestData(model.FunctionTypeIdentificationListData)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return msgCounter, nil
}

// return current values for Identification
func (i *Identification) GetValues() ([]IdentificationType, error) {
	rData := i.featureRemote.Data(model.FunctionTypeIdentificationListData)
	if rData == nil {
		return nil, ErrDataNotAvailable
	}

	data := rData.(*model.IdentificationListDataType)
	var resultSet []IdentificationType

	for _, item := range data.IdentificationData {
		if item.IdentificationValue == nil {
			continue
		}

		result := IdentificationType{
			Identifier: string(*item.IdentificationValue),
		}
		if item.IdentificationType != nil {
			result.Type = *item.IdentificationType
		}

		resultSet = append(resultSet, result)
	}

	return resultSet, nil
}