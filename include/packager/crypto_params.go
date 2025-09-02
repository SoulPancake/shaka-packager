// Copyright 2017 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package packager

import "fmt"

// KeyProvider represents encryption key providers. These provide keys to decrypt the content
// if the source content is encrypted, or used to encrypt the content.
type KeyProvider int

const (
	KeyProviderNone KeyProvider = iota
	KeyProviderRawKey
	KeyProviderWidevine
	KeyProviderPlayReady
)

// ProtectionSystem represents protection systems that handle decryption during playback.
// This affects the protection info that is stored in the content.
// Multiple protection systems can be combined using OR.
type ProtectionSystem uint16

const (
	ProtectionSystemNone ProtectionSystem = 0
	// The common key system from EME: https://goo.gl/s8RIhr
	ProtectionSystemCommon   ProtectionSystem = (1 << 0)
	ProtectionSystemWidevine ProtectionSystem = (1 << 1)
	ProtectionSystemPlayReady ProtectionSystem = (1 << 2)
	ProtectionSystemFairPlay ProtectionSystem = (1 << 3)
	ProtectionSystemMarlin   ProtectionSystem = (1 << 4)
)

// Bitwise operations for ProtectionSystem
func (a ProtectionSystem) Or(b ProtectionSystem) ProtectionSystem {
	return ProtectionSystem(uint16(a) | uint16(b))
}

func (a ProtectionSystem) And(b ProtectionSystem) ProtectionSystem {
	return ProtectionSystem(uint16(a) & uint16(b))
}

func (a ProtectionSystem) Not() ProtectionSystem {
	return ProtectionSystem(^uint16(a))
}

func (a ProtectionSystem) HasFlag(flag ProtectionSystem) bool {
	return (a & flag) == flag
}

// SigningKeyType represents the signing key type for Widevine signer.
type SigningKeyType int

const (
	SigningKeyTypeNone SigningKeyType = iota
	SigningKeyTypeAES
	SigningKeyTypeRSA
)

// WidevineSigner represents signer credential for Widevine license server.
type WidevineSigner struct {
	// Name of the signer / content provider.
	SignerName string

	// Specifies the signing key type, which determines whether AES or RSA key
	// are used to authenticate the signer. A type of 'None' is invalid.
	SigningKeyType SigningKeyType

	// AES signing credentials
	AES struct {
		// AES signing key.
		Key []byte
		// AES signing IV.
		IV []byte
	}

	// RSA signing credentials  
	RSA struct {
		// RSA signing private key.
		Key string
	}
}

// WidevineEncryptionParams represents Widevine encryption parameters.
type WidevineEncryptionParams struct {
	// Widevine license / key server URL.
	KeyServerURL string
	// Content identifier.
	ContentID []byte
	// The name of a stored policy, which specifies DRM content rights.
	Policy string
	// Signer credential for Widevine license / key server.
	Signer WidevineSigner
	// Group identifier, if present licenses will belong to this group.
	GroupID []byte
	// Enables entitlement license when set to true.
	EnableEntitlementLicense bool
}

// PlayReadyEncryptionParams represents PlayReady encryption parameters.
// KeyServerURL and ProgramIdentifier are required. The presence of
// other parameters may be necessary depends on server configuration.
type PlayReadyEncryptionParams struct {
	// PlayReady license / key server URL.
	KeyServerURL string
	// PlayReady program identifier.
	ProgramIdentifier string
	// Absolute path to the Certificate Authority file for the server cert in PEM format.
	CAFile string
	// Absolute path to client certificate file.
	ClientCertFile string
	// Absolute path to the private key file.
	ClientCertPrivateKeyFile string
	// Password to the private key file.
	ClientCertPrivateKeyPassword string
}

// StreamLabel is a type for stream labels.
type StreamLabel string

// KeyInfo represents key information for a stream.
type KeyInfo struct {
	KeyID []byte
	Key   []byte
	IV    []byte
}

// RawKeyParams represents raw key encryption/decryption parameters,
// i.e. with key parameters provided.
type RawKeyParams struct {
	// An optional initialization vector. If not provided, a random IV will be
	// generated. Note that this parameter should only be used during testing.
	// Not needed for decryption.
	IV []byte
	// Inject a custom PSSH or multiple concatenated PSSHs. If not provided,
	// a common system pssh will be generated.
	// Not needed for decryption.
	PSSH []byte

	// Defines the KeyInfo for the streams. An empty StreamLabel indicates the
	// default KeyInfo, which applies to all the StreamLabels not present in KeyMap.
	KeyMap map[StreamLabel]KeyInfo
}

// Protection scheme constants
const (
	ProtectionSchemeCENC = 0x63656E63
	ProtectionSchemeCBC1 = 0x63626331
	ProtectionSchemeCENS = 0x63656E73
	ProtectionSchemeCBCS = 0x63626373
)

// StreamType represents the type of encrypted stream.
type StreamType int

const (
	StreamTypeUnknown StreamType = iota
	StreamTypeVideo
	StreamTypeAudio
)

// EncryptedStreamAttributes represents encrypted stream information
// that is used to determine stream label.
type EncryptedStreamAttributes struct {
	StreamType StreamType

	// Video-specific attributes
	Video struct {
		Width     int
		Height    int
		FrameRate float32
		BitDepth  int
	}

	// Audio-specific attributes
	Audio struct {
		NumberOfChannels int
	}
}

// StreamLabelFunc assigns a stream label to the stream to be encrypted.
// Stream label is used to associate KeyPair with streams. Streams with the same
// stream label always use the same keyPair; Streams with different stream label
// could use the same or different KeyPairs.
// A default stream label function will be generated if not set.
type StreamLabelFunc func(streamAttributes EncryptedStreamAttributes) string

// EncryptionParams represents encryption parameters.
type EncryptionParams struct {
	// Specifies the key provider, which determines which key provider is used
	// and which encryption params is valid. 'None' means not to encrypt the streams.
	KeyProvider KeyProvider

	// Only one of the three fields is valid.
	Widevine  WidevineEncryptionParams
	PlayReady PlayReadyEncryptionParams
	RawKey    RawKeyParams

	// The protection systems to generate, multiple can be OR'd together.
	ProtectionSystems ProtectionSystem
	// Extra XML data to add to PlayReady data.
	PlayReadyExtraHeaderData string

	// Clear lead duration in seconds.
	ClearLeadInSeconds float64
	// The protection scheme: "cenc", "cens", "cbc1", "cbcs".
	ProtectionScheme uint32
	// The count of the encrypted blocks in the protection pattern, where each
	// block is of size 16-bytes. There are three common patterns
	// (crypt_byte_block:skip_byte_block): 1:9 (default), 5:5, 10:0.
	// Applies to video streams with "cbcs" and "cens" protection schemes only;
	// Ignored otherwise.
	CryptByteBlock uint8
	// The count of the unencrypted blocks in the protection pattern.
	// Applies to video streams with "cbcs" and "cens" protection schemes only;
	// Ignored otherwise.
	SkipByteBlock uint8
	// Crypto period duration in seconds. A positive value means key rotation is
	// enabled, the key provider must support key rotation in this case.
	// kNoKeyRotation = 0 means no key rotation.
	CryptoPeriodDurationInSeconds float64
	// Enable/disable subsample encryption for VP9.
	VP9SubsampleEncryption bool

	// Stream label function assigns a stream label to the stream to be encrypted.
	StreamLabelFunc StreamLabelFunc
}

// Default values for EncryptionParams
const (
	NoKeyRotation         = 0.0
	DefaultCryptByteBlock = 1
	DefaultSkipByteBlock  = 9
)

// NewEncryptionParams creates a new EncryptionParams with default values.
func NewEncryptionParams() EncryptionParams {
	return EncryptionParams{
		KeyProvider:                   KeyProviderNone,
		ProtectionSystems:             ProtectionSystemNone,
		ClearLeadInSeconds:            0,
		ProtectionScheme:              ProtectionSchemeCENC,
		CryptByteBlock:                DefaultCryptByteBlock,
		SkipByteBlock:                 DefaultSkipByteBlock,
		CryptoPeriodDurationInSeconds: NoKeyRotation,
		VP9SubsampleEncryption:        true,
	}
}

// WidevineDecryptionParams represents Widevine decryption parameters.
type WidevineDecryptionParams struct {
	// Widevine license / key server URL.
	KeyServerURL string
	// Signer credential for Widevine license / key server.
	Signer WidevineSigner
}

// DecryptionParams represents decryption parameters.
type DecryptionParams struct {
	// Specifies the key provider, which determines which key provider is used
	// and which decryption params is valid. 'None' means not to decrypt the streams.
	KeyProvider KeyProvider

	// Only one of the two fields is valid.
	Widevine WidevineDecryptionParams
	RawKey   RawKeyParams
}

// NewDecryptionParams creates a new DecryptionParams with default values.
func NewDecryptionParams() DecryptionParams {
	return DecryptionParams{
		KeyProvider: KeyProviderNone,
	}
}

// String methods for enums
func (kp KeyProvider) String() string {
	switch kp {
	case KeyProviderNone:
		return "None"
	case KeyProviderRawKey:
		return "RawKey" 
	case KeyProviderWidevine:
		return "Widevine"
	case KeyProviderPlayReady:
		return "PlayReady"
	default:
		return fmt.Sprintf("Unknown(%d)", int(kp))
	}
}

func (ps ProtectionSystem) String() string {
	var systems []string
	if ps&ProtectionSystemCommon != 0 {
		systems = append(systems, "Common")
	}
	if ps&ProtectionSystemWidevine != 0 {
		systems = append(systems, "Widevine")
	}
	if ps&ProtectionSystemPlayReady != 0 {
		systems = append(systems, "PlayReady")
	}
	if ps&ProtectionSystemFairPlay != 0 {
		systems = append(systems, "FairPlay")
	}
	if ps&ProtectionSystemMarlin != 0 {
		systems = append(systems, "Marlin")
	}
	if len(systems) == 0 {
		return "None"
	}
	result := ""
	for i, sys := range systems {
		if i > 0 {
			result += "|"
		}
		result += sys
	}
	return result
}

func (st StreamType) String() string {
	switch st {
	case StreamTypeUnknown:
		return "Unknown"
	case StreamTypeVideo:
		return "Video"
	case StreamTypeAudio:
		return "Audio"
	default:
		return fmt.Sprintf("Unknown(%d)", int(st))
	}
}