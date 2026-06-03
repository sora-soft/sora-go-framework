package provider

import (
	"testing"

	"github.com/sora-soft/sora-go-framework.git/pkg/discovery"
)

func TestEndpointChanged(t *testing.T) {
	tests := []struct {
		name     string
		a        discovery.EndpointMeta
		b        discovery.EndpointMeta
		expected bool
	}{
		{
			name: "identical",
			a: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:19347",
				Protocol: "tcp",
				Codecs:   []string{"json", "msgpack"},
			},
			b: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:19347",
				Protocol: "tcp",
				Codecs:   []string{"json", "msgpack"},
			},
			expected: false,
		},
		{
			name: "endpoint address changed",
			a: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:0",
				Protocol: "tcp",
				Codecs:   []string{"json"},
			},
			b: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:19347",
				Protocol: "tcp",
				Codecs:   []string{"json"},
			},
			expected: true,
		},
		{
			name: "protocol changed",
			a: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:8080",
				Protocol: "tcp",
				Codecs:   []string{"json"},
			},
			b: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:8080",
				Protocol: "ws",
				Codecs:   []string{"json"},
			},
			expected: true,
		},
		{
			name: "codecs changed - different length",
			a: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:8080",
				Protocol: "tcp",
				Codecs:   []string{"json"},
			},
			b: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:8080",
				Protocol: "tcp",
				Codecs:   []string{"json", "msgpack"},
			},
			expected: true,
		},
		{
			name: "codecs changed - same length different content",
			a: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:8080",
				Protocol: "tcp",
				Codecs:   []string{"json", "msgpack"},
			},
			b: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:8080",
				Protocol: "tcp",
				Codecs:   []string{"json", "protobuf"},
			},
			expected: true,
		},
		{
			name: "non-key fields changed - weight",
			a: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:8080",
				Protocol: "tcp",
				Codecs:   []string{"json"},
				Weight:   100,
			},
			b: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:8080",
				Protocol: "tcp",
				Codecs:   []string{"json"},
				Weight:   50,
			},
			expected: false,
		},
		{
			name: "non-key fields changed - labels",
			a: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:8080",
				Protocol: "tcp",
				Codecs:   []string{"json"},
				Labels:   map[string]string{"zone": "a"},
			},
			b: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:8080",
				Protocol: "tcp",
				Codecs:   []string{"json"},
				Labels:   map[string]string{"zone": "b"},
			},
			expected: false,
		},
		{
			name: "non-key fields changed - state",
			a: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:8080",
				Protocol: "tcp",
				Codecs:   []string{"json"},
				State:    2,
			},
			b: discovery.EndpointMeta{
				Endpoint: "192.168.10.158:8080",
				Protocol: "tcp",
				Codecs:   []string{"json"},
				State:    3,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := endpointChanged(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("endpointChanged() = %v, want %v", result, tt.expected)
			}
		})
	}
}
