package ltcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLtcpCollect(t *testing.T) {

	tests := []struct {
		datas        [][]byte
		expectLength int
	}{
		{
			datas: [][]byte{
				[]byte("Hello"),
				[]byte(", World"),
			},
			expectLength: 2,
		},
		{
			datas:        [][]byte{},
			expectLength: 1,
		},
	}

	for _, test := range tests {
		ltcp := NewLtcp(make(chan error, 1<<16))
		for _, bts := range test.datas {
			ltcp.Collect(bts)
		}
		ltcp.Package(1)
		assert.Equal(t, test.expectLength, ltcp.sendQueue.cnt)
	}
}
