package grpc_streaming

import (
	"fmt"
	"grpc-bidirectional-streaming/pkg/prometheus"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type MappingService struct {
	requestChanMap  cmap.ConcurrentMap[string, *chan any]
	responseChanMap cmap.ConcurrentMap[string, *chan any]
}

func NewMappingService() *MappingService {
	return &MappingService{
		requestChanMap:  cmap.New[*chan any](),
		responseChanMap: cmap.New[*chan any](),
	}
}

func (s *MappingService) Monitor() {
	prometheus.RequestChanNum.Set(float64(s.requestChanMap.Count()))
	prometheus.ResponseChanNum.Set(float64(s.responseChanMap.Count()))
}

func (s *MappingService) GetRequestChan(packageName string, funcName string, clientId string) (*chan any, error) {
	// Act
	requestChanIndex := s.getRequestChanIndex(packageName, funcName, clientId)
	requestChan, ok := s.requestChanMap.Get(requestChanIndex)
	if !ok {
		return nil, fmt.Errorf("request channel not found")
	}

	// Return
	return requestChan, nil
}

func (s *MappingService) SetRequestChan(packageName string, funcName string, clientId string, requestChan *chan any) {
	requestChanIndex := s.getRequestChanIndex(packageName, funcName, clientId)
	s.requestChanMap.Set(requestChanIndex, requestChan)
}

func (s *MappingService) RemoveRequestChan(packageName string, funcName string, clientId string) {
	requestChanIndex := s.getRequestChanIndex(packageName, funcName, clientId)
	s.requestChanMap.Remove(requestChanIndex)
}

func (s *MappingService) GetResponseChan(requestId string) (*chan any, error) {
	// Act
	responseChan, ok := s.responseChanMap.Get(requestId)
	if !ok {
		return nil, fmt.Errorf("response channel not found")
	}

	// Return
	return responseChan, nil
}

func (s *MappingService) SetResponseChan(requestId string, responseChan *chan any) {
	s.responseChanMap.Set(requestId, responseChan)
}

func (s *MappingService) RemoveResponseChan(requestId string) {
	s.responseChanMap.Remove(requestId)
}

func (s *MappingService) getRequestChanIndex(packageName string, funcName string, clientId string) string {
	return fmt.Sprintf("%s_%s_%s", packageName, funcName, clientId)
}
