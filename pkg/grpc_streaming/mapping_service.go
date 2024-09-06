package grpcstreaming

import (
	"fmt"
	"grpc-bidirectional-streaming/pkg/prometheus"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type MappingService struct {
	requestChanMap  cmap.ConcurrentMap[string, chan any]
	responseChanMap cmap.ConcurrentMap[string, chan any]
}

func NewMappingService() *MappingService {
	return &MappingService{
		requestChanMap:  cmap.New[chan any](),
		responseChanMap: cmap.New[chan any](),
	}
}

func (s *MappingService) Monitor() {
	prometheus.RequestChanNum.Set(float64(s.requestChanMap.Count()))
	prometheus.ResponseChanNum.Set(float64(s.responseChanMap.Count()))
}

func (s *MappingService) GetRequestChan(packageName string, funcName string, clientID string) (chan any, error) {
	// Act
	requestChanIndex := s.getRequestChanIndex(packageName, funcName, clientID)
	requestChan, ok := s.requestChanMap.Get(requestChanIndex)
	if !ok {
		return nil, fmt.Errorf("request channel not found")
	}

	// Return
	return requestChan, nil
}

func (s *MappingService) SetRequestChan(packageName string, funcName string, clientID string, requestChan chan any) {
	requestChanIndex := s.getRequestChanIndex(packageName, funcName, clientID)
	s.requestChanMap.Set(requestChanIndex, requestChan)
}

func (s *MappingService) RemoveRequestChan(packageName string, funcName string, clientID string) {
	requestChanIndex := s.getRequestChanIndex(packageName, funcName, clientID)
	s.requestChanMap.Remove(requestChanIndex)
}

func (s *MappingService) GetResponseChan(requestID string) (chan any, error) {
	// Act
	responseChan, ok := s.responseChanMap.Get(requestID)
	if !ok {
		return nil, fmt.Errorf("response channel not found")
	}

	// Return
	return responseChan, nil
}

func (s *MappingService) SetResponseChan(requestID string, responseChan chan any) {
	s.responseChanMap.Set(requestID, responseChan)
}

func (s *MappingService) RemoveResponseChan(requestID string) {
	s.responseChanMap.Remove(requestID)
}

func (s *MappingService) getRequestChanIndex(packageName string, funcName string, clientID string) string {
	return fmt.Sprintf("%s_%s_%s", packageName, funcName, clientID)
}
