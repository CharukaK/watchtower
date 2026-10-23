package isapi

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

type XMLVersion struct {
	XMLName xml.Name `xml:"StreamingChannelList"`
}

type StreamingChannelList struct {
	XMLVersion
	Version  string             `xml:"version,attr"`
	Channels []StreamingChannel `xml:"StreamingChannel"`
}

type StreamingChannel struct {
	Version   string                    `xml:"version,attr"`
	ID        int                       `xml:"id"`
	Name      string                    `xml:"channelName"`
	Enabled   bool                      `xml:"enabled"`
	Transport StreamingChannelTransport `xml:"Transport"`
	Video     StreamingChannelVideo     `xml:"Video"`
	Audio     StreamingChannelAudio     `xml:"Audio"`
}

type StreamingChannelTransport struct {
	ControlProtocols []StreamingControlProtocol `xml:"ControlProtocolList>ControlProtocol"`
}

type StreamingControlProtocol struct {
	Transport string `xml:"streamingTransport"`
}

type StreamingChannelVideo struct {
	Enabled           bool       `xml:"enabled"`
	InputChannelID    int        `xml:"videoInputChannelID"`
	CodecType         string     `xml:"videoCodecType"`
	ScanType          string     `xml:"videoScanType"`
	ResolutionWidth   int        `xml:"videoResolutionWidth"`
	ResolutionHeight  int        `xml:"videoResolutionHeight"`
	QualityControl    string     `xml:"videoQualityControlType"`
	FixedQuality      int        `xml:"fixedQuality"`
	VBRUpperCap       int        `xml:"vbrUpperCap"`
	VBRLowerCap       int        `xml:"vbrLowerCap"`
	MaxFrameRate      int        `xml:"maxFrameRate"`
	SnapshotImageType string     `xml:"snapShotImageType"`
	GOVLength         int        `xml:"GovLength"`
	SmartCodec        SmartCodec `xml:"SmartCodec"`
	VBRAverageCap     int        `xml:"vbrAverageCap"`
}

type SmartCodec struct {
	Enabled bool `xml:"enabled"`
}

type StreamingChannelAudio struct {
	Enabled         bool   `xml:"enabled"`
	InputChannelID  int    `xml:"audioInputChannelID"`
	CompressionType string `xml:"audioCompressionType"`
}

type CapabilityValue[T any] struct {
	Value   T      `xml:",chardata"`
	Options string `xml:"opt,attr"`
	Default T      `xml:"def,attr"`
	Min     int    `xml:"min,attr"`
	Max     int    `xml:"max,attr"`
	Name    string `xml:"name,attr"`
}

type StreamingChannelCapabilities struct {
	Version               string                                `xml:"version,attr"`
	ID                    int                                   `xml:"id"`
	Name                  CapabilityValue[string]               `xml:"channelName"`
	Enabled               CapabilityValue[bool]                 `xml:"enabled"`
	Transport             StreamingChannelTransportCapabilities `xml:"Transport"`
	Video                 StreamingChannelVideoCapabilities     `xml:"Video"`
	Audio                 StreamingChannelAudioCapabilities     `xml:"Audio"`
	IsSupportPictureByURL bool                                  `xml:"isSupportPictureByUrl"`
}

type StreamingChannelTransportCapabilities struct {
	ControlProtocols []StreamingControlProtocolCapabilities `xml:"ControlProtocolList>ControlProtocol"`
}

type StreamingControlProtocolCapabilities struct {
	Transport CapabilityValue[string] `xml:"streamingTransport"`
}

type StreamingChannelVideoCapabilities struct {
	Enabled           CapabilityValue[bool]   `xml:"enabled"`
	InputChannelID    CapabilityValue[int]    `xml:"videoInputChannelID"`
	CodecType         CapabilityValue[string] `xml:"videoCodecType"`
	ResolutionWidth   CapabilityValue[int]    `xml:"videoResolutionWidth"`
	ResolutionHeight  CapabilityValue[int]    `xml:"videoResolutionHeight"`
	ResolutionName    CapabilityValue[string] `xml:"videoResolutionName"`
	QualityControl    CapabilityValue[string] `xml:"videoQualityControlType"`
	FixedQuality      CapabilityValue[int]    `xml:"fixedQuality"`
	ConstantBitRate   CapabilityValue[int]    `xml:"constantBitRate"`
	VBRUpperCap       CapabilityValue[int]    `xml:"vbrUpperCap"`
	VBRLowerCap       CapabilityValue[int]    `xml:"vbrLowerCap"`
	MaxFrameRate      CapabilityValue[int]    `xml:"maxFrameRate"`
	SnapshotImageType CapabilityValue[string] `xml:"snapShotImageType"`
	SVC               StreamingChannelSVC     `xml:"SVC"`
}

type StreamingChannelSVC struct {
	Enabled CapabilityValue[bool]   `xml:"enabled"`
	Mode    CapabilityValue[string] `xml:"SVCMode"`
}

type StreamingChannelAudioCapabilities struct {
	Enabled         CapabilityValue[bool]   `xml:"enabled"`
	InputChannelID  CapabilityValue[int]    `xml:"audioInputChannelID"`
	CompressionType CapabilityValue[string] `xml:"audioCompressionType"`
}

func (c *Client) GetChannels(ctx context.Context) ([]StreamingChannel, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.do(ctx, http.MethodGet, "/Streaming/channels", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result StreamingChannelList

	// Prevent malformed firmware responses from consuming unlimited memory.
	reader := io.LimitReader(resp.Body, 1<<20)

	if err := xml.NewDecoder(reader).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode device information: %w", err)
	}

	return result.Channels, nil
}

func (c *Client) GetChannel(ctx context.Context, id string) (StreamingChannel, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if len(id) == 0 {
		return StreamingChannel{}, fmt.Errorf("invalid channel id")
	}

	resp, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/Streaming/channels/%s", id), nil)
	if err != nil {
		return StreamingChannel{}, err
	}
	defer resp.Body.Close()

	var result StreamingChannel

	// Prevent malformed firmware responses from consuming unlimited memory.
	reader := io.LimitReader(resp.Body, 1<<20)

	if err := xml.NewDecoder(reader).Decode(&result); err != nil {
		return StreamingChannel{}, fmt.Errorf("decode device information: %w", err)
	}

	return result, nil
}

func (c *Client) GetChannelCapabilities(ctx context.Context, id string) (StreamingChannelCapabilities, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if len(id) == 0 {
		return StreamingChannelCapabilities{}, fmt.Errorf("invalid channel id")
	}

	resp, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/Streaming/channels/%s/capabilities", id), nil)
	if err != nil {
		return StreamingChannelCapabilities{}, err
	}
	defer resp.Body.Close()

	var result StreamingChannelCapabilities

	// Prevent malformed firmware responses from consuming unlimited memory.
	reader := io.LimitReader(resp.Body, 1<<20)

	if err := xml.NewDecoder(reader).Decode(&result); err != nil {
		return StreamingChannelCapabilities{}, fmt.Errorf("decode channel capabilities: %w", err)
	}

	return result, nil
}
