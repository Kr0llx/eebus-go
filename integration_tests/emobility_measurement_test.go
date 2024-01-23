package integrationtests

import (
	"testing"
	"time"

	"github.com/enbility/eebus-go/features"
	"github.com/enbility/spine-go/api"
	"github.com/enbility/spine-go/model"
	"github.com/enbility/spine-go/spine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	m_descriptionListData_recv_reply_file_path           = "./testdata/m_descriptionListData_recv_reply.json"
	m_measurementListData_recv_notify_file_path          = "./testdata/m_measurementListData_recv_notify.json"
	ec_parameterdescriptionlistdata_recv_reply_file_path = "./testdata/ec_parameterDescriptionListData_recv_reply.json"
)

func TestEmobilityMeasurementSuite(t *testing.T) {
	suite.Run(t, new(EmobilityMeasurementSuite))
}

type EmobilityMeasurementSuite struct {
	suite.Suite

	sut         api.DeviceLocalInterface
	localEntity api.EntityLocalInterface

	measurement          *features.Measurement
	electricalconnection *features.ElectricalConnection

	remoteSki string

	remoteDevice api.DeviceRemoteInterface
	writeHandler *WriteMessageHandler
}

func (s *EmobilityMeasurementSuite) BeforeTest(suiteName, testName string) {
	s.sut = spine.NewDeviceLocal("TestBrandName", "TestDeviceModel", "TestSerialNumber", "TestDeviceCode",
		"TestDeviceAddress", model.DeviceTypeTypeEnergyManagementSystem, model.NetworkManagementFeatureSetTypeSmart, time.Second*4)
	s.localEntity = spine.NewEntityLocal(s.sut, model.EntityTypeTypeCEM, spine.NewAddressEntityType([]uint{1}))
	s.sut.AddEntity(s.localEntity)

	f := spine.NewFeatureLocal(1, s.localEntity, model.FeatureTypeTypeElectricalConnection, model.RoleTypeClient)
	s.localEntity.AddFeature(f)
	f = spine.NewFeatureLocal(2, s.localEntity, model.FeatureTypeTypeMeasurement, model.RoleTypeClient)
	s.localEntity.AddFeature(f)

	s.remoteSki = "TestRemoteSki"

	s.writeHandler = &WriteMessageHandler{}
	_ = s.sut.SetupRemoteDevice(s.remoteSki, s.writeHandler)
	s.remoteDevice = s.sut.RemoteDeviceForSki(s.remoteSki)

	initialCommunication(s.T(), s.remoteDevice, s.writeHandler)
}

func (s *EmobilityMeasurementSuite) TestGetValuesPerPhaseForScope() {
	remoteEntity := s.sut.RemoteDeviceForSki(s.remoteSki).Entity([]model.AddressEntityType{1, 1})
	assert.NotNil(s.T(), remoteEntity)

	var err error
	s.measurement, err = features.NewMeasurement(model.RoleTypeClient, model.RoleTypeServer, s.localEntity, remoteEntity)
	assert.Nil(s.T(), err)

	s.electricalconnection, err = features.NewElectricalConnection(model.RoleTypeClient, model.RoleTypeServer, s.localEntity, remoteEntity)
	assert.Nil(s.T(), err)

	// Act
	msgCounter, _ := s.remoteDevice.HandleSpineMesssage(loadFileData(s.T(), ec_parameterdescriptionlistdata_recv_reply_file_path))
	waitForAck(s.T(), msgCounter, s.writeHandler)

	msgCounter, _ = s.remoteDevice.HandleSpineMesssage(loadFileData(s.T(), m_descriptionListData_recv_reply_file_path))
	waitForAck(s.T(), msgCounter, s.writeHandler)

	msgCounter, _ = s.remoteDevice.HandleSpineMesssage(loadFileData(s.T(), m_measurementListData_recv_notify_file_path))
	waitForAck(s.T(), msgCounter, s.writeHandler)

	measurement := model.MeasurementTypeTypeCurrent
	commodity := model.CommodityTypeTypeElectricity
	scope := model.ScopeTypeTypeACCurrent
	data, err := s.measurement.GetValuesForTypeCommodityScope(measurement, commodity, scope)

	// Assert
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), data)
	assert.Equal(s.T(), 1, len(data))
	assert.Equal(s.T(), 5.0, data[0].Value.GetValue())

	measurement = model.MeasurementTypeTypePower
	scope = model.ScopeTypeTypeACPower
	data, err = s.measurement.GetValuesForTypeCommodityScope(measurement, commodity, scope)

	// Assert
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), data)
	assert.Equal(s.T(), 1, len(data))
	assert.Equal(s.T(), 1185.0, data[0].Value.GetValue())

	measurement = model.MeasurementTypeTypeEnergy
	scope = model.ScopeTypeTypeCharge
	data, err = s.measurement.GetValuesForTypeCommodityScope(measurement, commodity, scope)

	// Assert
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), data)
	assert.Equal(s.T(), 1, len(data))
	assert.Equal(s.T(), 1825.0, data[0].Value.GetValue())
}
