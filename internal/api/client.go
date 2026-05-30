package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type SystemStatus struct {
	APIVersion string `json:"apiVersion"`
	Hostname   string `json:"hostname"`
	UpTime     int64  `json:"upTime"`
}

type DeviceInfo struct {
	Connected int    `json:"connected"`
	Family    int    `json:"family"`
	ID        string `json:"id"`
	Model     int    `json:"model"`
	Name      string `json:"name"`
	Port      struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"port"`
	UserName string `json:"userName"`
	UserProd string `json:"userProd"`
}

type DeviceListResponse struct {
	Devices []DeviceInfo `json:"devices"`
}

type DeviceStatusResponse struct {
	Device DeviceStatus `json:"device"`
}

type DeviceStatus struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Model  int    `json:"model"`
	Family int    `json:"family"`
	Status struct {
		Connected int `json:"connected"`
		SyncInput int `json:"syncInput"`
	} `json:"status"`
	Vars DeviceVars `json:"vars"`
}

type DeviceVars struct {
	CBattery        float64 `json:"cBattery"`
	FOutput         float64 `json:"fOutput"`
	IOutput         float64 `json:"iOutput"`
	ICBattery       int     `json:"icBattery"`
	LedBlue         int     `json:"ledBlue"`
	LedGreen        int     `json:"ledGreen"`
	LedRed          int     `json:"ledRed"`
	NominalFOutput  float64 `json:"nominalFOutput"`
	NominalPOutput  float64 `json:"nominalPOutput"`
	NominalVBattery float64 `json:"nominalVBattery"`
	NominalVInput   float64 `json:"nominalVInput"`
	NominalVOutput  float64 `json:"nominalVOutput"`
	POutput         float64 `json:"pOutput"`
	Temperature     float64 `json:"temperature"`
	VBattery        float64 `json:"vBattery"`
	VInput          float64 `json:"vInput"`
	VOutput         float64 `json:"vOutput"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetSystemStatus(ctx context.Context) (*SystemStatus, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/mon/1.1")
	if err != nil {
		return nil, fmt.Errorf("get system status: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get system status: status %d", resp.StatusCode)
	}
	var status SystemStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decode system status: %w", err)
	}
	return &status, nil
}

func (c *Client) GetDevices(ctx context.Context) ([]DeviceInfo, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/mon/1.1/device")
	if err != nil {
		return nil, fmt.Errorf("get devices: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get devices: status %d", resp.StatusCode)
	}
	var list DeviceListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("decode devices: %w", err)
	}
	return list.Devices, nil
}

func (c *Client) GetDeviceStatus(ctx context.Context, deviceID string) (*DeviceStatus, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/mon/1.1/device/" + deviceID)
	if err != nil {
		return nil, fmt.Errorf("get device status: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get device status: status %d", resp.StatusCode)
	}
	var status DeviceStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decode device status: %w", err)
	}
	return &status.Device, nil
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.GetSystemStatus(ctx)
	return err
}
