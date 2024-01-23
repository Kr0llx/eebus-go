package service

import (
	"crypto/tls"
	"testing"
	"time"

	"github.com/enbility/eebus-go/api"
	"github.com/enbility/eebus-go/mocks"
	shipapi "github.com/enbility/ship-go/api"
	"github.com/enbility/ship-go/cert"
	"github.com/enbility/ship-go/logging"
	shipmocks "github.com/enbility/ship-go/mocks"
	"github.com/enbility/spine-go/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

type ServiceSuite struct {
	suite.Suite

	config *api.Configuration

	sut *Service

	serviceReader *mocks.ServiceReaderInterface
	conHub        *shipmocks.HubInterface
	logging       *shipmocks.LoggingInterface
}

func (s *ServiceSuite) BeforeTest(suiteName, testName string) {
	s.serviceReader = mocks.NewServiceReaderInterface(s.T())

	s.conHub = shipmocks.NewHubInterface(s.T())

	s.logging = shipmocks.NewLoggingInterface(s.T())

	certificate := tls.Certificate{}
	s.config, _ = api.NewConfiguration(
		"vendor", "brand", "model", "serial", model.DeviceTypeTypeEnergyManagementSystem,
		[]model.EntityTypeType{model.EntityTypeTypeCEM}, 4729, certificate, 230.0, time.Second*4)

	s.sut = NewService(s.config, s.serviceReader)
}

func (s *ServiceSuite) Test_EEBUSHandler() {
	testSki := "test"

	entry := &shipapi.MdnsEntry{
		Ski: testSki,
	}

	entries := []*shipapi.MdnsEntry{entry}
	s.serviceReader.EXPECT().VisibleRemoteServicesUpdated(mock.Anything, mock.Anything).Return()
	s.sut.VisibleMDNSRecordsUpdated(entries)

	s.serviceReader.EXPECT().RemoteSKIConnected(mock.Anything, mock.Anything).Return()
	s.sut.RemoteSKIConnected(testSki)

	s.serviceReader.EXPECT().RemoteSKIDisconnected(mock.Anything, mock.Anything).Return()
	s.sut.RemoteSKIDisconnected(testSki)

	s.serviceReader.EXPECT().ServiceShipIDUpdate(mock.Anything, mock.Anything).Return()
	s.sut.ServiceShipIDUpdate(testSki, "shipid")

	s.serviceReader.EXPECT().ServicePairingDetailUpdate(mock.Anything, mock.Anything).Return()
	detail := &shipapi.ConnectionStateDetail{}
	s.sut.ServicePairingDetailUpdate(testSki, detail)

	s.serviceReader.EXPECT().AllowWaitingForTrust(mock.Anything).Return(true)
	result := s.sut.AllowWaitingForTrust(testSki)
	assert.Equal(s.T(), true, result)

}

func (s *ServiceSuite) Test_ConnectionsHub() {
	testSki := "test"

	s.sut.connectionsHub = s.conHub

	s.conHub.EXPECT().PairingDetailForSki(mock.Anything).Return(nil)
	s.sut.PairingDetailForSki(testSki)

	s.conHub.EXPECT().StartBrowseMdnsSearch().Return()
	s.sut.StartBrowseMdnsEntries()

	s.conHub.EXPECT().StopBrowseMdnsSearch().Return()
	s.sut.StopBrowseMdnsEntries()

	s.conHub.EXPECT().ServiceForSKI(mock.Anything).Return(nil)
	details := s.sut.RemoteServiceForSKI(testSki)
	assert.Nil(s.T(), details)

	s.conHub.EXPECT().RegisterRemoteSKI(mock.Anything, mock.Anything).Return()
	s.sut.RegisterRemoteSKI(testSki, true)

	s.conHub.EXPECT().InitiatePairingWithSKI(mock.Anything).Return()
	s.sut.InitiatePairingWithSKI(testSki)

	s.conHub.EXPECT().CancelPairingWithSKI(mock.Anything).Return()
	s.sut.CancelPairingWithSKI(testSki)

	s.conHub.EXPECT().DisconnectSKI(mock.Anything, mock.Anything).Return()
	s.sut.DisconnectSKI(testSki, "reason")
}

func (s *ServiceSuite) Test_SetLogging() {
	s.sut.SetLogging(nil)
	assert.Equal(s.T(), &logging.NoLogging{}, logging.Log())

	s.sut.SetLogging(s.logging)
	assert.Equal(s.T(), s.logging, logging.Log())

	s.sut.SetLogging(&logging.NoLogging{})
	assert.Equal(s.T(), &logging.NoLogging{}, logging.Log())
}

func (s *ServiceSuite) Test_Setup() {

	err := s.sut.Setup()
	assert.NotNil(s.T(), err)

	certificate, err := cert.CreateCertificate("unit", "org", "de", "cn")
	assert.Nil(s.T(), err)
	s.config.SetCertificate(certificate)

	err = s.sut.Setup()
	assert.Nil(s.T(), err)

	s.sut.connectionsHub = s.conHub
	s.conHub.EXPECT().Start()
	s.sut.Start()

	time.Sleep(time.Millisecond * 200)

	s.conHub.EXPECT().Shutdown()
	s.sut.Shutdown()

	device := s.sut.LocalDevice()
	assert.NotNil(s.T(), device)
}
