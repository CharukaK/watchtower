package isapi

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

type DeviceInfo struct {
	XMLName              xml.Name `xml:"DeviceInfo"`
	Name                 string   `xml:"deviceName"`
	DeviceID             string   `xml:"deviceID"`
	Model                string   `xml:"model"`
	SerialNumber         string   `xml:"serialNumber"`
	MACAddress           string   `xml:"macAddress"`
	Firmware             string   `xml:"firmwareVersion"`
	FirmwareReleasedDate string   `xml:"firmwareReleasedDate"`
	EncoderVersion       string   `xml:"encoderVersion"`
	EncoderReleasedDate  string   `xml:"encoderReleasedDate"`
	Type                 string   `xml:"deviceType"`
	TelecontrolID        int      `xml:"telecontrolID"`
	HardwareVersion      string   `xml:"hardwareVersion"`
	Manufacturer         string   `xml:"manufacturer"`
	Status               string   `xml:"deviceStatus"`
	Substream            bool     `xml:"subChannelEnabled"`
	ThirdStream          bool     `xml:"thrChannelEnabled"`
}

func (c *Client) GetDeviceInfo(ctx context.Context) (DeviceInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.do(ctx, http.MethodGet, "System/deviceInfo", nil)
	if err != nil {
		return DeviceInfo{}, err
	}
	defer resp.Body.Close()

	var result DeviceInfo

	// Prevent malformed firmware responses from consuming unlimited memory.
	reader := io.LimitReader(resp.Body, 1<<20)

	if err := xml.NewDecoder(reader).Decode(&result); err != nil {
		return DeviceInfo{}, fmt.Errorf("decode device information: %w", err)
	}

	return result, nil
}
