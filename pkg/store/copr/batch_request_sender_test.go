package copr

import (
	"context"
	"testing"
	"time"

	"github.com/pingcap/errors"
	"github.com/pingcap/kvproto/pkg/coprocessor"
	"github.com/pingcap/kvproto/pkg/metapb"
	"github.com/pingcap/tidb/pkg/config"
	tikverr "github.com/tikv/client-go/v2/error"
	"github.com/tikv/client-go/v2/oracle"
	"github.com/tikv/client-go/v2/tikv"
	"github.com/tikv/client-go/v2/tikvrpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockRegionCache is a mock implementation of RegionCache for testing.
type MockRegionCache struct {
	*tikv.RegionCache
}

func (m *MockRegionCache) OnSendFailForBatchRegions(bo *Backoffer, store *tikv.Store, regionInfos []RegionInfo, reloadRegion bool, err error) {
	// Mock implementation for testing
}

// MockClient is a mock implementation of tikv.Client for testing.
type MockClient struct {
	tikv.Client
	sendRequestFunc func(ctx context.Context, addr string, req *tikvrpc.Request, timeout time.Duration) (*tikvrpc.Response, error)
}

func (m *MockClient) SendRequest(ctx context.Context, addr string, req *tikvrpc.Request, timeout time.Duration) (*tikvrpc.Response, error) {
	return m.sendRequestFunc(ctx, addr, req, timeout)
}

// TestSendReqToAddr tests the SendReqToAddr method.
func TestSendReqToAddr(t *testing.T) {
	// Setup
	mockCache := &MockRegionCache{}
	mockClient := &MockClient{}
	oracle := oracle.NewMockOracle()
	sender := NewRegionBatchRequestSender(mockCache, mockClient, oracle, true)

	// Test cases
	tests := []struct {
		name          string
		mockResponse  *tikvrpc.Response
		mockError     error
		expectedRetry bool
		expectedErr   error
	}{
		{
			name:          "successful request",
			mockResponse:  &tikvrpc.Response{},
			mockError:     nil,
			expectedRetry: false,
			expectedErr:   nil,
		},
		{
			name:          "context canceled",
			mockResponse:  nil,
			mockError:     context.Canceled,
			expectedRetry: false,
			expectedErr:   context.Canceled,
		},
		{
			name:          "stale epoch error",
			mockResponse:  nil,
			mockError:     &tikverr.ErrRegion{StaleEpoch: true},
			expectedRetry: true,
			expectedErr:   nil,
		},
		{
			name:          "other error",
			mockResponse:  nil,
			mockError:     errors.New("other error"),
			expectedRetry: true,
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the client's SendRequest function
			mockClient.sendRequestFunc = func(ctx context.Context, addr string, req *tikvrpc.Request, timeout time.Duration) (*tikvrpc.Response, error) {
				return tt.mockResponse, tt.mockError
			}

			// Create a backoffer and RPC context
			bo := NewBackoffer(context.Background(), 100)
			rpcCtx := &tikv.RPCContext{
				Meta: &metapb.Region{},
				Peer: &metapb.Peer{},
				Addr: "127.0.0.1:4000",
			}

			// Call SendReqToAddr
			resp, retry, cancel, err := sender.SendReqToAddr(bo, rpcCtx, []RegionInfo{}, &tikvrpc.Request{}, time.Second)

			// Verify the results
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, errors.Cause(err))
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedRetry, retry)
			if resp != nil {
				assert.Equal(t, tt.mockResponse, resp)
			}
			cancel() // Ensure the cancel function is called
		})
	}
}

// TestOnSendFailForBatchRegions tests the onSendFailForBatchRegions method.
func TestOnSendFailForBatchRegions(t *testing.T) {
	// Setup
	mockCache := &MockRegionCache{}
	mockClient := &MockClient{}
	oracle := oracle.NewMockOracle()
	sender := NewRegionBatchRequestSender(mockCache, mockClient, oracle, true)

	// Test cases
	tests := []struct {
		name          string
		err           error
		expectedError error
	}{
		{
			name:          "context canceled",
			err:           context.Canceled,
			expectedError: context.Canceled,
		},
		{
			name:          "stale epoch error",
			err:           &tikverr.ErrRegion{StaleEpoch: true},
			expectedError: nil,
		},
		{
			name:          "other error",
			err:           errors.New("other error"),
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a backoffer and RPC context
			bo := NewBackoffer(context.Background(), 100)
			rpcCtx := &tikv.RPCContext{
				Meta: &metapb.Region{},
				Peer: &metapb.Peer{},
				Addr: "127.0.0.1:4000",
			}

			// Call onSendFailForBatchRegions
			err := sender.onSendFailForBatchRegions(bo, rpcCtx, []RegionInfo{}, tt.err)

			// Verify the results
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, errors.Cause(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}