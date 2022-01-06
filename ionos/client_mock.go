package ionos

import (
	"context"
	"fmt"
)

type MockClient struct {
	config     *Config
	txtRecords map[string]string
}

func (e MockClient) SetConfig(ctx context.Context, config *Config) {
	e.config = config
}

func (e MockClient) GetZoneIdByName(ctx context.Context, name string) (string, error) {
	// Unmarshall response
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
		fmt.Printf(zone.Name)
	}

	return "", fmt.Errorf("unable tu find zone %v", name)
}

func (e MockClient) GetRecordIdByName(ctx context.Context, zoneId string, recordName string) (string, error) {

	for key, _ := range e.txtRecords {
		if key == recordName+"."+zoneId {
			return key, nil
		}
	}

	return "", fmt.Errorf("unable tu find record %v", recordName)
}

func (e MockClient) GetRecordById(ctx context.Context, zoneId string, recordId string) (string, error) {
	value, ok := e.txtRecords[recordId]
	if ok {
		return value, nil
	} else {
		return "", fmt.Errorf("unable tu find record %v", recordId)
	}
}

func (e MockClient) AddRecord(ctx context.Context, zoneId string, records RecordCreateRequest) error {

	for _, record := range records {
		e.txtRecords[record.Name+"."+zoneId] = record.Content
	}

	return nil
}

func (e MockClient) DeleteRecord(ctx context.Context, zoneId string, recordId string) error {
	delete(e.txtRecords, recordId)
	return nil
}

func (e MockClient) GetZoneById(ctx context.Context, id string) (string, error) {
	panic("implement me")
}

func NewMockClient() Client {
	return &MockClient{txtRecords: make(map[string]string)}
}
