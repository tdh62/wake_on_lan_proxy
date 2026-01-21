package main

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestParseMACAddress(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid MAC with colons",
			input:   "AA:BB:CC:DD:EE:FF",
			want:    "aabbccddeeff",
			wantErr: false,
		},
		{
			name:    "Valid MAC with hyphens",
			input:   "AA-BB-CC-DD-EE-FF",
			want:    "aabbccddeeff",
			wantErr: false,
		},
		{
			name:    "Valid MAC without separators",
			input:   "AABBCCDDEEFF",
			want:    "aabbccddeeff",
			wantErr: false,
		},
		{
			name:    "Valid MAC lowercase",
			input:   "aa:bb:cc:dd:ee:ff",
			want:    "aabbccddeeff",
			wantErr: false,
		},
		{
			name:    "Valid MAC with spaces",
			input:   "AA BB CC DD EE FF",
			want:    "aabbccddeeff",
			wantErr: false,
		},
		{
			name:    "Invalid MAC - too short",
			input:   "AA:BB:CC:DD:EE",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid MAC - too long",
			input:   "AA:BB:CC:DD:EE:FF:00",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid MAC - invalid characters",
			input:   "GG:HH:II:JJ:KK:LL",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Empty MAC",
			input:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMACAddress(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMACAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotHex := hex.EncodeToString(got)
				if gotHex != tt.want {
					t.Errorf("parseMACAddress() = %v, want %v", gotHex, tt.want)
				}
			}
		})
	}
}

func TestCreateMagicPacket(t *testing.T) {
	tests := []struct {
		name string
		mac  []byte
	}{
		{
			name: "Standard MAC address",
			mac:  []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},
		},
		{
			name: "All zeros MAC",
			mac:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "All ones MAC",
			mac:  []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := createMagicPacket(tt.mac)

			// 检查包长度
			if len(packet) != 102 {
				t.Errorf("Magic packet length = %d, want 102", len(packet))
			}

			// 检查前6个字节是否都是0xFF
			for i := 0; i < 6; i++ {
				if packet[i] != 0xFF {
					t.Errorf("Byte %d = 0x%02X, want 0xFF", i, packet[i])
				}
			}

			// 检查后面是否重复了16次MAC地址
			for i := 0; i < 16; i++ {
				offset := 6 + i*6
				macSlice := packet[offset : offset+6]
				if !bytes.Equal(macSlice, tt.mac) {
					t.Errorf("MAC repetition %d = %v, want %v", i, macSlice, tt.mac)
				}
			}
		})
	}
}

func TestCreateMagicPacketStructure(t *testing.T) {
	mac := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	packet := createMagicPacket(mac)

	// 验证魔术包的完整结构
	expected := make([]byte, 102)

	// 前6个字节为0xFF
	for i := 0; i < 6; i++ {
		expected[i] = 0xFF
	}

	// 后面重复16次MAC地址
	for i := 0; i < 16; i++ {
		copy(expected[6+i*6:], mac)
	}

	if !bytes.Equal(packet, expected) {
		t.Errorf("Magic packet structure mismatch")
		t.Logf("Got:      %x", packet)
		t.Logf("Expected: %x", expected)
	}
}

func BenchmarkParseMACAddress(b *testing.B) {
	macAddr := "AA:BB:CC:DD:EE:FF"
	for i := 0; i < b.N; i++ {
		parseMACAddress(macAddr)
	}
}

func BenchmarkCreateMagicPacket(b *testing.B) {
	mac := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	for i := 0; i < b.N; i++ {
		createMagicPacket(mac)
	}
}
