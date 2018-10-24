package types

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/ethereumproject/go-ethereum/rlp"
)

func encodeReceipt(r *Receipt) ([]byte, error) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	err := r.EncodeRLP(writer)
	writer.Flush()

	return buf.Bytes(), err
}

func TestEIP658RLPRoundTrip1(t *testing.T) {
	// EIP-658 enabled - PostState is nil, Status is present
	r1 := NewReceipt(nil, big.NewInt(4095))
	r1.Status = TxFailure

	rlpData, err := encodeReceipt(r1)
	if err != nil {
		t.Error("unexpected error", err)
	}

	var r2 Receipt
	err = r2.DecodeRLP(rlp.NewStream(bytes.NewReader(rlpData), 0))
	if err != nil {
		t.Error("decoding error:", err)
	}

	if r1.Status != r2.Status {
		t.Errorf("invalid status: expected %v, got %v", r1.Status, r2.Status)
	}

}

func TestEIP658RLPRoundTrip2(t *testing.T) {
	// EIP-658 disabled - PostState AND Status are present in Receipt
	r1 := NewReceipt(common.Hash{}.Bytes(), big.NewInt(4095))
	for i := 0; i < len(r1.PostState); i++ {
		r1.PostState[i] = byte(i)
	}
	r1.Status = TxSuccess

	rlpData, err := encodeReceipt(r1)
	if err != nil {
		t.Error("unexpected error", err)
	}

	var r2 Receipt
	r2.DecodeRLP(rlp.NewStream(bytes.NewReader(rlpData), 0))

	if !bytes.Equal(r1.PostState, r2.PostState) {
		t.Errorf("invalid PostState: expected %v, got %v", r1.PostState, r2.PostState)
	}
	if r2.Status != TxStatusUnknown {
		t.Errorf("invalid Status: expected 0xFF, got %v", r2.Status)
	}
}

func TestInvalidReceiptsEncoding(t *testing.T) {
	// case 1: invalid PostState
	r := NewReceipt(make([]byte, 7), big.NewInt(4095))
	_, err := encodeReceipt(r)
	if err == nil {
		t.Error("error was expected")
	} else if err.Error() != fmt.Sprintf(errfInvalidStateLen, len(r.PostState)) {
		t.Error("invalid error message:", err)
	}

	// case 2: no PostState (EIP-658), unknown transaction status
	r = NewReceipt(nil, big.NewInt(4095))
	_, err = encodeReceipt(r)
	if err == nil {
		t.Error("error was expected")
	} else if err.Error() != fmt.Sprintf(errfNoStateNorStatus) {
		t.Error("invalid error message:", err)
	}
}

func TestInvalidReceiptsDecoding(t *testing.T) {
	// This is the valid hex-encoded RLP from TestEIP658RLPRoundTrip1
	// f9010801820fffb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0

	// change comments legend:
	// v - changed value
	// l - changed lenght

	// Lets change status to invalid value - 0x22
	//                 vv
	invalid1 := "f9010822820fffb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0"

	// Lets change status to ivalid value - 0xEE (value over 0x79 are encoded differently)
	// Note, that lenght also needs to be changed
	//                lvvvv
	invalid2 := "f9010981EE820fffb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0"

	// Lets change status to ivalid value - 0xFF (this is special case, because 0xFF is used internally to denote unknown
	// status, but it's not supported to use such Status in consensus Receipt)
	// Note, that lenght also needs to be changed
	//                lvvvv
	invalid3 := "f9010981FF820fffb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0"

	// Try to use something bigger - []byte{0x01, 0x01}
	// Note, that lenght also needs to be changed
	//                lvvvvvv
	invalid4 := "f9010A820101820fffb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0"

	// Invalid encoding of TxFailure - 0x00 instead of 0x80
	//                 vv
	invalid5 := "f9010800820fffb9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0"

	testCases := []struct {
		name             string
		rlpHex           string
		expectedErrorMsg string
	}{
		{
			"Status=0x22",
			invalid1,
			fmt.Sprintf(errfInvalidStatus, 0x22),
		},
		{
			"Status=0xEE",
			invalid2,
			fmt.Sprintf(errfInvalidStatus, 0xEE),
		},
		{
			"Status=0xFF(TxStatusUnknown)",
			invalid3,
			fmt.Sprintf(errfInvalidStatus, 0xFF),
		},
		{
			"Status=0x0101",
			invalid4,
			fmt.Sprintf(errfInvalidStateOrStatus, hex.EncodeToString([]byte{0x01, 0x01})),
		},
		{
			"Status=0x00",
			invalid5,
			fmt.Sprintf(errfInvalidStatus, 0x00),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(tt *testing.T) {
			rlpData, _ := hex.DecodeString(testCase.rlpHex)
			var r Receipt
			err := r.DecodeRLP(rlp.NewStream(bytes.NewReader(rlpData), 0))
			if err == nil {
				tt.Error("error was expected")
			} else {
				if err.Error() != testCase.expectedErrorMsg {
					tt.Error("invalid error message:", err)
				}
			}
		})
	}
}