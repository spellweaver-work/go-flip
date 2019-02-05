package flip

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	clockwork "github.com/keybase/clockwork"
	"github.com/stretchr/testify/require"
	"testing"
)

type testDealersHelper struct {
	clock clockwork.FakeClock
}

func newTestDealersHelper() *testDealersHelper {
	return &testDealersHelper{clock: clockwork.NewFakeClock()}
}

func (t *testDealersHelper) Clock() clockwork.Clock {
	return t.clock
}

func (t *testDealersHelper) CLogf(ctx context.Context, fmtString string, args ...interface{}) {
	fmt.Printf(fmtString, args...)
}

func randBytes(i int) []byte {
	ret := make([]byte, i)
	rand.Read(ret[:])
	return ret
}

func newTestUserDevice() UserDevice {
	return UserDevice{
		U: randBytes(6),
		D: randBytes(6),
	}
}

func newGameMessageWrappedEncoded(t *testing.T, u UserDevice, g GameID, b GameMessageBody) GameMessageWrappedEncoded {
	v1 := GameMessageV1{
		GameID: g,
		Body:   b,
	}
	msg := NewGameMessageWithV1(v1)
	raw, err := msgpackEncode(msg)
	require.NoError(t, err)
	return GameMessageWrappedEncoded{
		Header: u,
		Body:   base64.StdEncoding.EncodeToString(raw),
	}
}

func TestDealer(t *testing.T) {
	dh := newTestDealersHelper()
	dealer := NewDealer(dh)
	ctx := context.Background()
	go func() {
		dealer.Run(ctx)
	}()

	leader := newTestUserDevice()
	params := NewFlipParametersWithInts([]FlipParametersInt{NewFlipParametersIntWithBool()})
	start := Start{
		StartTime:            ToTime(dh.clock.Now()),
		CommitmentWindowMsec: 5 * 1000,
		RevealWindowMsec:     5 * 1000,
		Params:               params,
	}

	gameID := GenerateGameID()
	body := NewGameMessageBodyWithStart(start)
	gmwe := newGameMessageWrappedEncoded(t, leader, gameID, body)

	dealer.MessageCh() <- gmwe
}
