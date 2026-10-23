package isapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Streaming/channels/101" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/Streaming/channels/101")
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8" ?>
<StreamingChannel version="1.0" xmlns="http://www.hikvision.com/ver20/XMLSchema">
<id>101</id>
<channelName>101</channelName>
<enabled>true</enabled>
<Transport><ControlProtocolList><ControlProtocol><streamingTransport>RTSP</streamingTransport></ControlProtocol></ControlProtocolList></Transport>
<Video>
<enabled>true</enabled><videoInputChannelID>1</videoInputChannelID><videoCodecType>H.265</videoCodecType>
<videoResolutionWidth>960</videoResolutionWidth><videoResolutionHeight>1080</videoResolutionHeight>
<videoQualityControlType>VBR</videoQualityControlType><fixedQuality>60</fixedQuality>
<vbrUpperCap>1024</vbrUpperCap><vbrLowerCap>32</vbrLowerCap><maxFrameRate>1500</maxFrameRate>
<snapShotImageType>JPEG</snapShotImageType><GovLength>15</GovLength><SmartCodec><enabled>false</enabled></SmartCodec><vbrAverageCap>832</vbrAverageCap>
</Video>
<Audio><enabled>true</enabled><audioInputChannelID>1</audioInputChannelID><audioCompressionType>G.711ulaw</audioCompressionType></Audio>
</StreamingChannel>`))
	}))
	defer server.Close()

	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := &Client{baseURL: baseURL, http: server.Client()}

	channel, err := client.GetChannel(context.Background(), "101")
	if err != nil {
		t.Fatal(err)
	}
	if channel.Version != "1.0" || channel.ID != 101 || channel.Name != "101" || !channel.Enabled {
		t.Fatalf("channel = %+v", channel)
	}
	if channel.Video.InputChannelID != 1 || channel.Video.CodecType != "H.265" || channel.Video.ResolutionWidth != 960 || channel.Video.ResolutionHeight != 1080 {
		t.Fatalf("video = %+v", channel.Video)
	}
	if len(channel.Transport.ControlProtocols) != 1 || channel.Transport.ControlProtocols[0].Transport != "RTSP" {
		t.Fatalf("transport = %+v", channel.Transport)
	}
	if !channel.Audio.Enabled || channel.Audio.CompressionType != "G.711ulaw" {
		t.Fatalf("audio = %+v", channel.Audio)
	}
}

func TestGetChannelCapabilities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Streaming/channels/101/capabilities" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/Streaming/channels/101/capabilities")
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8" ?>
<StreamingChannel version="1.0" xmlns="http://www.hikvision.com/ver20/XMLSchema">
<id>101</id>
<channelName name="any string less than 31 words">101</channelName>
<enabled opt="true,false" def="true">true</enabled>
<Transport><ControlProtocolList><ControlProtocol><streamingTransport opt="RTSP">RTSP</streamingTransport></ControlProtocol></ControlProtocolList></Transport>
<Video>
<enabled opt="true">true</enabled><videoInputChannelID opt="1">1</videoInputChannelID><videoCodecType opt="H.264,H.265">H.265</videoCodecType>
<videoResolutionWidth opt="960,640,960,704,640,352">960</videoResolutionWidth><videoResolutionHeight opt="1080,720,576,576,480,288">1080</videoResolutionHeight>
<videoResolutionName opt="none,720P Lite,none,none,none,none">none</videoResolutionName><videoQualityControlType opt="CBR,VBR">VBR</videoQualityControlType>
<fixedQuality opt="20,30,45,60,75,90">60</fixedQuality><constantBitRate min="32" max="4096">0</constantBitRate>
<vbrUpperCap min="32" max="4096">1024</vbrUpperCap><vbrLowerCap opt="32">32</vbrLowerCap><maxFrameRate opt="1500,1200,1000,800,600,400,200,100">1500</maxFrameRate>
<snapShotImageType opt="JPEG">JPEG</snapShotImageType><SVC><enabled opt="true, false">false</enabled><SVCMode opt="manual,auto">manual</SVCMode></SVC>
</Video>
<Audio><enabled opt="true, false">true</enabled><audioInputChannelID opt="1">1</audioInputChannelID><audioCompressionType opt="G.711ulaw">G.711ulaw</audioCompressionType></Audio>
<isSupportPictureByUrl>true</isSupportPictureByUrl>
</StreamingChannel>`))
	}))
	defer server.Close()

	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := &Client{baseURL: baseURL, http: server.Client()}

	capabilities, err := client.GetChannelCapabilities(context.Background(), "101")
	if err != nil {
		t.Fatal(err)
	}
	if capabilities.Version != "1.0" || capabilities.ID != 101 || capabilities.Name.Value != "101" || capabilities.Name.Name != "any string less than 31 words" {
		t.Fatalf("capabilities = %+v", capabilities)
	}
	if !capabilities.Enabled.Value || capabilities.Enabled.Options != "true,false" || !capabilities.Enabled.Default {
		t.Fatalf("enabled = %+v", capabilities.Enabled)
	}
	if capabilities.Video.CodecType.Value != "H.265" || capabilities.Video.CodecType.Options != "H.264,H.265" {
		t.Fatalf("codec = %+v", capabilities.Video.CodecType)
	}
	if capabilities.Video.ConstantBitRate.Min != 32 || capabilities.Video.ConstantBitRate.Max != 4096 {
		t.Fatalf("constant bit rate = %+v", capabilities.Video.ConstantBitRate)
	}
	if capabilities.Video.SVC.Mode.Value != "manual" || capabilities.Video.SVC.Mode.Options != "manual,auto" {
		t.Fatalf("SVC = %+v", capabilities.Video.SVC)
	}
	if !capabilities.Audio.Enabled.Value || capabilities.Audio.CompressionType.Value != "G.711ulaw" || !capabilities.IsSupportPictureByURL {
		t.Fatalf("audio or picture support = %+v, %t", capabilities.Audio, capabilities.IsSupportPictureByURL)
	}
}
