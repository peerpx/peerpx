package cryptobox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const RSAPubKeyOK = `-----BEGIN PUBLIC KEY-----
MIIBCgKCAQEAqVIsY/YRF/+Y3R5vHi8EsNr4fTxFQiYtDCHKj1Jd6eTV+LpxZesn
+jspCUXEID0bowbUXly+QkBsA3ZBFOAE4vmd+XQ3ukt+aHHWJnJVpZjrMScDIYrJ
RENXAMyW4yZ1tnL66efm5/qsYypqOEICLr27A0+yIwlJ4vjlziy+rEwFihdJKorv
RBCAiYBUgio7l9Y+Oo0kqd/ZL8DtBHYqsSyTcRcHL/s/O2Ktyxo7cUsvelmTClS2
zjCJHAVwlnaPzFzVuG9WYTT9j1bU8JInAhxDSOylJKJoCtUrx1vJp+yF4N/JtXGZ
+oP/W8u+1TQl1G54j0MFyalZjtEzEpe+RQIDAQAB
-----END PUBLIC KEY-----`

const RSAPubKeyKO = `-----BEGIN PRIVATE KEY-----
MIIBCgKCAQEAqVIsY/YRF/+Y3R5vHi8EsNr4fTxFQiYtDCHKj1Jd6eTV+LpxZesn
+jspCUXEID0bowbUXly+QkBsA3ZBFOAE4vmd+XQ3ukt+aHHWJnJVpZjrMScDIYrJ
RENXAMyW4yZ1tnL66efm5/qsYypqOEICLr27A0+yIwlJ4vjlziy+rEwFihdJKorv
RBCAiYBUgio7l9Y+Oo0kqd/ZL8DtBHYqsSyTcRcHL/s/O2Ktyxo7cUsvelmTClS2
zjCJHAVwlnaPzFzVuG9WYTT9j1bU8JInAhxDSOylJKJoCtUrx1vJp+yF4N/JtXGZ
+oP/W8u+1TQl1G54j0MFyalZjtEzEpe+RQIDAQAB
-----END PUBLIC KEY-----`

const RSAPubKeyKOKO = `-----BEGIN PUBLIC KEY-----
MIIBCgKCAQEAqVIsY/YRF/+Y3R5vHi8EsNr4fTxFQiYtDCHKj1Jd6eTV+LpxZesn
+jspCUXEItUrx1vJp+yF4N/JtXGZ
+oP/W8u+1TQl1G54j0MFyalZjtEzEpe+RQIDAQAB
-----END PUBLIC KEY-----`

const MagicKeyOK = "RSA.qVIsY_YRF_-Y3R5vHi8EsNr4fTxFQiYtDCHKj1Jd6eTV-LpxZesn-jspCUXEID0bowbUXly-QkBsA3ZBFOAE4vmd-XQ3ukt-aHHWJnJVpZjrMScDIYrJRENXAMyW4yZ1tnL66efm5_qsYypqOEICLr27A0-yIwlJ4vjlziy-rEwFihdJKorvRBCAiYBUgio7l9Y-Oo0kqd_ZL8DtBHYqsSyTcRcHL_s_O2Ktyxo7cUsvelmTClS2zjCJHAVwlnaPzFzVuG9WYTT9j1bU8JInAhxDSOylJKJoCtUrx1vJp-yF4N_JtXGZ-oP_W8u-1TQl1G54j0MFyalZjtEzEpe-RQ.AQAB"
const MagicKeyKO = "ASR.qVIsY_YRF_-Y3R5vHi8EsNr4fTxFQiYtDCHKj1Jd6eTV-LpxZesn-jspCUXEID0bowbUXly-QkBsA3ZBFOAE4vmd-XQ3ukt-aHHWJnJVpZjrMScDIYrJRENXAMyW4yZ1tnL66efm5_qsYypqOEICLr27A0-yIwlJ4vjlziy-rEwFihdJKorvRBCAiYBUgio7l9Y-Oo0kqd_ZL8DtBHYqsSyTcRcHL_s_O2Ktyxo7cUsvelmTClS2zjCJHAVwlnaPzFzVuG9WYTT9j1bU8JInAhxDSOylJKJoCtUrx1vJp-yF4N_JtXGZ-oP_W8u-1TQl1G54j0MFyalZjtEzEpe-RQ.AQAB"
const MagicKeyKO2p = "RSA.OylJKJoCtUrx1vJp-yF4N_JtXGZ-oP_W8u-1TQl1G54j0MFyalZjtEzEpe-RQAQAB"
const MagicKeyKOp1 = "RSA.0.AQAB"
const MagicKeyKOp2 = "RSA.qVIsY_YRF_-Y3R5vHi8EsNr4fTxFQiYtDCHKj1Jd6eTV-LpxZesn-jspCUXEID0bowbUXly-QkBsA3ZBFOAE4vmd-XQ3ukt-aHHWJnJVpZjrMScDIYrJRENXAMyW4yZ1tnL66efm5_qsYypqOEICLr27A0-yIwlJ4vjlziy-rEwFihdJKorvRBCAiYBUgio7l9Y-Oo0kqd_ZL8DtBHYqsSyTcRcHL_s_O2Ktyxo7cUsvelmTClS2zjCJHAVwlnaPzFzVuG9WYTT9j1bU8JInAhxDSOylJKJoCtUrx1vJp-yF4N_JtXGZ-oP_W8u-1TQl1G54j0MFyalZjtEzEpe-RQ.0"

func TestRSAGenerateKeysAsPemStr(t *testing.T) {
	// ok
	_, _, err := RSAGenerateKeysAsPemStr()
	assert.NoError(t, err)
}

func TestRSAGetMagicKey(t *testing.T) {
	// ko
	_, err := RSAGetMagicKey(RSAPubKeyKO)
	assert.Error(t, err)

	// ko
	_, err = RSAGetMagicKey(RSAPubKeyKOKO)
	assert.Error(t, err)

	// ok
	magicKey, err := RSAGetMagicKey(RSAPubKeyOK)
	if assert.NoError(t, err) {
		assert.Equal(t, "RSA.qVIsY_YRF_-Y3R5vHi8EsNr4fTxFQiYtDCHKj1Jd6eTV-LpxZesn-jspCUXEID0bowbUXly-QkBsA3ZBFOAE4vmd-XQ3ukt-aHHWJnJVpZjrMScDIYrJRENXAMyW4yZ1tnL66efm5_qsYypqOEICLr27A0-yIwlJ4vjlziy-rEwFihdJKorvRBCAiYBUgio7l9Y-Oo0kqd_ZL8DtBHYqsSyTcRcHL_s_O2Ktyxo7cUsvelmTClS2zjCJHAVwlnaPzFzVuG9WYTT9j1bU8JInAhxDSOylJKJoCtUrx1vJp-yF4N_JtXGZ-oP_W8u-1TQl1G54j0MFyalZjtEzEpe-RQ.AQAB", magicKey)
	}
}

func TestRSAParseMagicKey(t *testing.T) {
	// ko
	_, err := RSAParseMagicKey(MagicKeyKO)
	assert.Error(t, err)
	_, err = RSAParseMagicKey(MagicKeyKO2p)
	assert.Error(t, err)
	_, err = RSAParseMagicKey(MagicKeyKOp1)
	assert.Error(t, err)
	_, err = RSAParseMagicKey(MagicKeyKOp2)
	assert.Error(t, err)

	// ok
	_, err = RSAParseMagicKey(MagicKeyOK)
	assert.NoError(t, err)

}
