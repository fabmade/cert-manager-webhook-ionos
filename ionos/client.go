package ionos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
)

type Client interface {
	GetZoneIdByName(ctx context.Context, name string) (string, error)
	GetZoneById(ctx context.Context, id string) (string, error)
	SetConfig(ctx context.Context, config *Config)
	AddRecord(ctx context.Context, zoneId string, records RecordCreateRequest) error
	GetRecordIdByName(ctx context.Context, zoneId string, recordName string) (string, error)
	GetRecordById(ctx context.Context, zoneId string, recordId string) (string, error)
	DeleteRecord(ctx context.Context, zoneId string, recordId string) error
}

type IonosClient struct {
	config *Config
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

func (e *IonosClient) callDNSApi(url string, method string, body io.Reader, config *Config) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to execute request %v", err)
	}

	if config == nil {
		return []byte{}, fmt.Errorf("config missing")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", config.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			klog.Fatal(err)
		}
	}()

	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return respBody, nil
	}

	text := "Error calling API status:" + resp.Status + " url: " + url + " method: " + method
	klog.Error(text)
	return nil, errors.New(text)
}

func (e *IonosClient) GetZoneIdByName(ctx context.Context, name string) (string, error) {

	url := e.config.ApiUrl + "/zones"

	zoneRecords, err := e.callDNSApi(url, "GET", nil, e.config)

	if err != nil {
		return "", fmt.Errorf("unable to get zone info %v", err)
	}

	// Unmarshall response
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

	return "", fmt.Errorf("unable tu find zone %v", name)
}

func (e *IonosClient) GetRecordIdByName(ctx context.Context, zoneId string, recordName string) (string, error) {
	url := e.config.ApiUrl + "/zones/" + zoneId + "?recordName=" + recordName + "&recordType=TXT"

	// Get all DNS records
	dnsRecords, err := e.callDNSApi(url, "GET", nil, e.config)

	if err != nil {
		return "", fmt.Errorf("unable to get DNS records %v", err)
	}

	// Unmarshall response
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
	panic("implement me")
}

func (e *IonosClient) AddRecord(ctx context.Context, zoneId string, records RecordCreateRequest) error {
	url := e.config.ApiUrl + "/zones/" + zoneId + "/records"

	jsonString, err := json.Marshal(records)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		fmt.Println(string(jsonString))
	}
	_, err = e.callDNSApi(url, "POST", bytes.NewBuffer(jsonString), e.config)
	return err
}

func (e *IonosClient) DeleteRecord(ctx context.Context, zoneId string, recordId string) error {
	url := e.config.ApiUrl + "/zones/" + zoneId + "/records/" + recordId
	_, err := e.callDNSApi(url, "DELETE", nil, e.config)
	return err
}

func (e *IonosClient) GetZoneById(ctx context.Context, id string) (string, error) {
	panic("implement me")
}

func NewClient() Client {
	return &IonosClient{}
}
