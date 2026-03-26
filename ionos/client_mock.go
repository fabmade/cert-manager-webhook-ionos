package ionos

import (
	"context"
	"fmt"
	"sync"
)

type MockClient struct {
	config     *Config
	txtRecords map[string]string
	mu         sync.RWMutex
}

func (e *MockClient) SetConfig(ctx context.Context, config *Config) {
	e.config = config
}

func (e *MockClient) GetZoneIdByName(ctx context.Context, name string) (string, error) {
	zones := ZoneResponse{}
	zones = append(zones, Zone{
		Id:   name + ".",
		Name: name,
		Type: "NATIVE",
	})

	for _, zone := range zones {
		if zone.Name == name {
			return zone.Id, nil
		}
	}

	return "", fmt.Errorf("unable to find zone %v", name)
}

func (e *MockClient) GetRecordIdByName(ctx context.Context, zoneId string, recordName string) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	fqdn := recordName + "."
	for key := range e.txtRecords {
		if key == fqdn {
			return key, nil
		}
	}

	return "", fmt.Errorf("unable to find record %v", recordName)
}

func (e *MockClient) GetRecordById(ctx context.Context, zoneId string, recordId string) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	value, ok := e.txtRecords[recordId]
	if ok {
		return value, nil
	}
	return "", fmt.Errorf("unable to find record %v", recordId)
}

func (e *MockClient) AddRecord(ctx context.Context, zoneId string, records RecordCreateRequest) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, record := range records {
		// Store with FQDN (trailing dot) to match DNS query format
		e.txtRecords[record.Name+"."] = record.Content
	}

	return nil
}

func (e *MockClient) DeleteRecord(ctx context.Context, zoneId string, recordId string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.txtRecords, recordId)
	return nil
}

func NewMockClient() Client {
	return &MockClient{txtRecords: make(map[string]string)}
}
