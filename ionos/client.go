package ionos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client interface {
	SetConfig(ctx context.Context, config *Config)
	GetZoneIdByName(ctx context.Context, name string) (string, error)
	AddRecord(ctx context.Context, zoneId string, records RecordCreateRequest) error
	GetRecordIdByName(ctx context.Context, zoneId string, recordName string) (string, error)
	GetRecordById(ctx context.Context, zoneId string, recordId string) (string, error)
	DeleteRecord(ctx context.Context, zoneId string, recordId string) error
}

type IonosClient struct {
	config     *Config
	httpClient *http.Client
}

type RecordResponse struct {
	Name    string   `json:"name"`
	Id      string   `json:"id"`
	Type    string   `json:"type"`
	Records []Record `json:"records"`
}

type RecordCreateRequest []RecordCreate

type ZoneResponse []Zone

type RecordCreate struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	Ttl      int    `json:"ttl"`
	Prio     int    `json:"prio"`
	Disabled bool   `json:"disabled"`
}

type Record struct {
	Name       string `json:"name"`
	RootName   string `json:"rootName"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	ChangeDate string `json:"changeDate"`
	Ttl        int    `json:"ttl"`
	Disabled   bool   `json:"disabled"`
	Id         string `json:"id"`
}

type Zone struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (e *IonosClient) SetConfig(ctx context.Context, config *Config) {
	e.config = config
}

func (e *IonosClient) callDNSApi(ctx context.Context, url string, method string, body io.Reader) ([]byte, error) {
	if e.config == nil {
		return nil, fmt.Errorf("config missing")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", e.config.ApiKey)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return respBody, nil
	}

	return nil, fmt.Errorf("API error status:%s url:%s method:%s", resp.Status, url, method)
}

func (e *IonosClient) GetZoneIdByName(ctx context.Context, name string) (string, error) {
	url := e.config.ApiUrl + "/zones"

	zoneRecords, err := e.callDNSApi(ctx, url, "GET", nil)

	if err != nil {
		return "", fmt.Errorf("unable to get zone info %v", err)
	}

	zones := ZoneResponse{}
	readErr := json.Unmarshal(zoneRecords, &zones)

	if readErr != nil {
		return "", fmt.Errorf("unable to unmarshal response %v", readErr)
	}

	for _, zone := range zones {
		if zone.Name == name {
			return zone.Id, nil
		}
	}

	return "", fmt.Errorf("unable to find zone %v", name)
}

func (e *IonosClient) GetRecordIdByName(ctx context.Context, zoneId string, recordName string) (string, error) {
	url := e.config.ApiUrl + "/zones/" + zoneId + "?recordName=" + recordName + "&recordType=TXT"

	dnsRecords, err := e.callDNSApi(ctx, url, "GET", nil)

	if err != nil {
		return "", fmt.Errorf("unable to get DNS records %v", err)
	}

	records := RecordResponse{}
	readErr := json.Unmarshal(dnsRecords, &records)

	if readErr != nil {
		return "", fmt.Errorf("unable to unmarshal response %v", readErr)
	}

	if len(records.Records) == 0 {
		return "", fmt.Errorf("unable to find record")
	}

	recordId := records.Records[0].Id

	return recordId, nil
}

func (e *IonosClient) GetRecordById(ctx context.Context, zoneId string, recordId string) (string, error) {
	return "", fmt.Errorf("GetRecordById not implemented")
}

func (e *IonosClient) AddRecord(ctx context.Context, zoneId string, records RecordCreateRequest) error {
	url := e.config.ApiUrl + "/zones/" + zoneId + "/records"

	jsonString, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("unable to marshal record request: %v", err)
	}
	_, err = e.callDNSApi(ctx, url, "POST", bytes.NewBuffer(jsonString))
	return err
}

func (e *IonosClient) DeleteRecord(ctx context.Context, zoneId string, recordId string) error {
	url := e.config.ApiUrl + "/zones/" + zoneId + "/records/" + recordId
	_, err := e.callDNSApi(ctx, url, "DELETE", nil)
	return err
}

func NewClient() Client {
	return &IonosClient{
		httpClient: &http.Client{},
	}
}
