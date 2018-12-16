// AUTO GENERATED FILE (by membufc proto compiler v0.0.21)
package leanhelix

import (
	"bytes"
	"fmt"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/membuffers/go"
)

/////////////////////////////////////////////////////////////////////////////
// message LeanHelixMessageContent

// reader

type LeanHelixMessageContent struct {
	// SignedHeader SignedHeader
	// Sender SenderSignature

	// internal
	// implements membuffers.Message
	_message membuffers.InternalMessage
}

func (x *LeanHelixMessageContent) String() string {
	if x == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{SignedHeader:%s,Sender:%s,}", x.StringSignedHeader(), x.StringSender())
}

var _LeanHelixMessageContent_Scheme = []membuffers.FieldType{membuffers.TypeMessage, membuffers.TypeMessage}
var _LeanHelixMessageContent_Unions = [][]membuffers.FieldType{}

func LeanHelixMessageContentReader(buf []byte) *LeanHelixMessageContent {
	x := &LeanHelixMessageContent{}
	x._message.Init(buf, membuffers.Offset(len(buf)), _LeanHelixMessageContent_Scheme, _LeanHelixMessageContent_Unions)
	return x
}

func (x *LeanHelixMessageContent) IsValid() bool {
	return x._message.IsValid()
}

func (x *LeanHelixMessageContent) Raw() []byte {
	return x._message.RawBuffer()
}

func (x *LeanHelixMessageContent) Equal(y *LeanHelixMessageContent) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return bytes.Equal(x.Raw(), y.Raw())
}

func (x *LeanHelixMessageContent) SignedHeader() *SignedHeader {
	b, s := x._message.GetMessage(0)
	return SignedHeaderReader(b[:s])
}

func (x *LeanHelixMessageContent) RawSignedHeader() []byte {
	return x._message.RawBufferForField(0, 0)
}

func (x *LeanHelixMessageContent) RawSignedHeaderWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(0, 0)
}

func (x *LeanHelixMessageContent) StringSignedHeader() string {
	return x.SignedHeader().String()
}

func (x *LeanHelixMessageContent) Sender() *SenderSignature {
	b, s := x._message.GetMessage(1)
	return SenderSignatureReader(b[:s])
}

func (x *LeanHelixMessageContent) RawSender() []byte {
	return x._message.RawBufferForField(1, 0)
}

func (x *LeanHelixMessageContent) RawSenderWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(1, 0)
}

func (x *LeanHelixMessageContent) StringSender() string {
	return x.Sender().String()
}

// builder

type LeanHelixMessageContentBuilder struct {
	SignedHeader *SignedHeaderBuilder
	Sender       *SenderSignatureBuilder

	// internal
	// implements membuffers.Builder
	_builder               membuffers.InternalBuilder
	_overrideWithRawBuffer []byte
}

func (w *LeanHelixMessageContentBuilder) Write(buf []byte) (err error) {
	if w == nil {
		return
	}
	w._builder.NotifyBuildStart()
	defer w._builder.NotifyBuildEnd()
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	if w._overrideWithRawBuffer != nil {
		return w._builder.WriteOverrideWithRawBuffer(buf, w._overrideWithRawBuffer)
	}
	w._builder.Reset()
	err = w._builder.WriteMessage(buf, w.SignedHeader)
	if err != nil {
		return
	}
	err = w._builder.WriteMessage(buf, w.Sender)
	if err != nil {
		return
	}
	return nil
}

func (w *LeanHelixMessageContentBuilder) HexDump(prefix string, offsetFromStart membuffers.Offset) (err error) {
	if w == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	w._builder.Reset()
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "LeanHelixMessageContent.SignedHeader", w.SignedHeader)
	if err != nil {
		return
	}
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "LeanHelixMessageContent.Sender", w.Sender)
	if err != nil {
		return
	}
	return nil
}

func (w *LeanHelixMessageContentBuilder) GetSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	return w._builder.GetSize()
}

func (w *LeanHelixMessageContentBuilder) CalcRequiredSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	w.Write(nil)
	return w._builder.GetSize()
}

func (w *LeanHelixMessageContentBuilder) Build() *LeanHelixMessageContent {
	buf := make([]byte, w.CalcRequiredSize())
	if w.Write(buf) != nil {
		return nil
	}
	return LeanHelixMessageContentReader(buf)
}

func LeanHelixMessageContentBuilderFromRaw(raw []byte) *LeanHelixMessageContentBuilder {
	return &LeanHelixMessageContentBuilder{_overrideWithRawBuffer: raw}
}

/////////////////////////////////////////////////////////////////////////////
// message SignedHeader

// reader

type SignedHeader struct {
	// MessageType MessageType

	// internal
	// implements membuffers.Message
	_message membuffers.InternalMessage
}

func (x *SignedHeader) String() string {
	if x == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{MessageType:%s,}", x.StringMessageType())
}

var _SignedHeader_Scheme = []membuffers.FieldType{membuffers.TypeUint16}
var _SignedHeader_Unions = [][]membuffers.FieldType{}

func SignedHeaderReader(buf []byte) *SignedHeader {
	x := &SignedHeader{}
	x._message.Init(buf, membuffers.Offset(len(buf)), _SignedHeader_Scheme, _SignedHeader_Unions)
	return x
}

func (x *SignedHeader) IsValid() bool {
	return x._message.IsValid()
}

func (x *SignedHeader) Raw() []byte {
	return x._message.RawBuffer()
}

func (x *SignedHeader) Equal(y *SignedHeader) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return bytes.Equal(x.Raw(), y.Raw())
}

func (x *SignedHeader) MessageType() MessageType {
	return MessageType(x._message.GetUint16(0))
}

func (x *SignedHeader) RawMessageType() []byte {
	return x._message.RawBufferForField(0, 0)
}

func (x *SignedHeader) MutateMessageType(v MessageType) error {
	return x._message.SetUint16(0, uint16(v))
}

func (x *SignedHeader) StringMessageType() string {
	return x.MessageType().String()
}

// builder

type SignedHeaderBuilder struct {
	MessageType MessageType

	// internal
	// implements membuffers.Builder
	_builder               membuffers.InternalBuilder
	_overrideWithRawBuffer []byte
}

func (w *SignedHeaderBuilder) Write(buf []byte) (err error) {
	if w == nil {
		return
	}
	w._builder.NotifyBuildStart()
	defer w._builder.NotifyBuildEnd()
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	if w._overrideWithRawBuffer != nil {
		return w._builder.WriteOverrideWithRawBuffer(buf, w._overrideWithRawBuffer)
	}
	w._builder.Reset()
	w._builder.WriteUint16(buf, uint16(w.MessageType))
	return nil
}

func (w *SignedHeaderBuilder) HexDump(prefix string, offsetFromStart membuffers.Offset) (err error) {
	if w == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	w._builder.Reset()
	w._builder.HexDumpUint16(prefix, offsetFromStart, "SignedHeader.MessageType", uint16(w.MessageType))
	return nil
}

func (w *SignedHeaderBuilder) GetSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	return w._builder.GetSize()
}

func (w *SignedHeaderBuilder) CalcRequiredSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	w.Write(nil)
	return w._builder.GetSize()
}

func (w *SignedHeaderBuilder) Build() *SignedHeader {
	buf := make([]byte, w.CalcRequiredSize())
	if w.Write(buf) != nil {
		return nil
	}
	return SignedHeaderReader(buf)
}

func SignedHeaderBuilderFromRaw(raw []byte) *SignedHeaderBuilder {
	return &SignedHeaderBuilder{_overrideWithRawBuffer: raw}
}

/////////////////////////////////////////////////////////////////////////////
// message PreprepareContent

// reader

type PreprepareContent struct {
	// SignedHeader BlockRef
	// Sender SenderSignature

	// internal
	// implements membuffers.Message
	_message membuffers.InternalMessage
}

func (x *PreprepareContent) String() string {
	if x == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{SignedHeader:%s,Sender:%s,}", x.StringSignedHeader(), x.StringSender())
}

var _PreprepareContent_Scheme = []membuffers.FieldType{membuffers.TypeMessage, membuffers.TypeMessage}
var _PreprepareContent_Unions = [][]membuffers.FieldType{}

func PreprepareContentReader(buf []byte) *PreprepareContent {
	x := &PreprepareContent{}
	x._message.Init(buf, membuffers.Offset(len(buf)), _PreprepareContent_Scheme, _PreprepareContent_Unions)
	return x
}

func (x *PreprepareContent) IsValid() bool {
	return x._message.IsValid()
}

func (x *PreprepareContent) Raw() []byte {
	return x._message.RawBuffer()
}

func (x *PreprepareContent) Equal(y *PreprepareContent) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return bytes.Equal(x.Raw(), y.Raw())
}

func (x *PreprepareContent) SignedHeader() *BlockRef {
	b, s := x._message.GetMessage(0)
	return BlockRefReader(b[:s])
}

func (x *PreprepareContent) RawSignedHeader() []byte {
	return x._message.RawBufferForField(0, 0)
}

func (x *PreprepareContent) RawSignedHeaderWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(0, 0)
}

func (x *PreprepareContent) StringSignedHeader() string {
	return x.SignedHeader().String()
}

func (x *PreprepareContent) Sender() *SenderSignature {
	b, s := x._message.GetMessage(1)
	return SenderSignatureReader(b[:s])
}

func (x *PreprepareContent) RawSender() []byte {
	return x._message.RawBufferForField(1, 0)
}

func (x *PreprepareContent) RawSenderWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(1, 0)
}

func (x *PreprepareContent) StringSender() string {
	return x.Sender().String()
}

// builder

type PreprepareContentBuilder struct {
	SignedHeader *BlockRefBuilder
	Sender       *SenderSignatureBuilder

	// internal
	// implements membuffers.Builder
	_builder               membuffers.InternalBuilder
	_overrideWithRawBuffer []byte
}

func (w *PreprepareContentBuilder) Write(buf []byte) (err error) {
	if w == nil {
		return
	}
	w._builder.NotifyBuildStart()
	defer w._builder.NotifyBuildEnd()
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	if w._overrideWithRawBuffer != nil {
		return w._builder.WriteOverrideWithRawBuffer(buf, w._overrideWithRawBuffer)
	}
	w._builder.Reset()
	err = w._builder.WriteMessage(buf, w.SignedHeader)
	if err != nil {
		return
	}
	err = w._builder.WriteMessage(buf, w.Sender)
	if err != nil {
		return
	}
	return nil
}

func (w *PreprepareContentBuilder) HexDump(prefix string, offsetFromStart membuffers.Offset) (err error) {
	if w == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	w._builder.Reset()
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "PreprepareContent.SignedHeader", w.SignedHeader)
	if err != nil {
		return
	}
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "PreprepareContent.Sender", w.Sender)
	if err != nil {
		return
	}
	return nil
}

func (w *PreprepareContentBuilder) GetSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	return w._builder.GetSize()
}

func (w *PreprepareContentBuilder) CalcRequiredSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	w.Write(nil)
	return w._builder.GetSize()
}

func (w *PreprepareContentBuilder) Build() *PreprepareContent {
	buf := make([]byte, w.CalcRequiredSize())
	if w.Write(buf) != nil {
		return nil
	}
	return PreprepareContentReader(buf)
}

func PreprepareContentBuilderFromRaw(raw []byte) *PreprepareContentBuilder {
	return &PreprepareContentBuilder{_overrideWithRawBuffer: raw}
}

/////////////////////////////////////////////////////////////////////////////
// message PrepareContent

// reader

type PrepareContent struct {
	// SignedHeader BlockRef
	// Sender SenderSignature

	// internal
	// implements membuffers.Message
	_message membuffers.InternalMessage
}

func (x *PrepareContent) String() string {
	if x == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{SignedHeader:%s,Sender:%s,}", x.StringSignedHeader(), x.StringSender())
}

var _PrepareContent_Scheme = []membuffers.FieldType{membuffers.TypeMessage, membuffers.TypeMessage}
var _PrepareContent_Unions = [][]membuffers.FieldType{}

func PrepareContentReader(buf []byte) *PrepareContent {
	x := &PrepareContent{}
	x._message.Init(buf, membuffers.Offset(len(buf)), _PrepareContent_Scheme, _PrepareContent_Unions)
	return x
}

func (x *PrepareContent) IsValid() bool {
	return x._message.IsValid()
}

func (x *PrepareContent) Raw() []byte {
	return x._message.RawBuffer()
}

func (x *PrepareContent) Equal(y *PrepareContent) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return bytes.Equal(x.Raw(), y.Raw())
}

func (x *PrepareContent) SignedHeader() *BlockRef {
	b, s := x._message.GetMessage(0)
	return BlockRefReader(b[:s])
}

func (x *PrepareContent) RawSignedHeader() []byte {
	return x._message.RawBufferForField(0, 0)
}

func (x *PrepareContent) RawSignedHeaderWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(0, 0)
}

func (x *PrepareContent) StringSignedHeader() string {
	return x.SignedHeader().String()
}

func (x *PrepareContent) Sender() *SenderSignature {
	b, s := x._message.GetMessage(1)
	return SenderSignatureReader(b[:s])
}

func (x *PrepareContent) RawSender() []byte {
	return x._message.RawBufferForField(1, 0)
}

func (x *PrepareContent) RawSenderWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(1, 0)
}

func (x *PrepareContent) StringSender() string {
	return x.Sender().String()
}

// builder

type PrepareContentBuilder struct {
	SignedHeader *BlockRefBuilder
	Sender       *SenderSignatureBuilder

	// internal
	// implements membuffers.Builder
	_builder               membuffers.InternalBuilder
	_overrideWithRawBuffer []byte
}

func (w *PrepareContentBuilder) Write(buf []byte) (err error) {
	if w == nil {
		return
	}
	w._builder.NotifyBuildStart()
	defer w._builder.NotifyBuildEnd()
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	if w._overrideWithRawBuffer != nil {
		return w._builder.WriteOverrideWithRawBuffer(buf, w._overrideWithRawBuffer)
	}
	w._builder.Reset()
	err = w._builder.WriteMessage(buf, w.SignedHeader)
	if err != nil {
		return
	}
	err = w._builder.WriteMessage(buf, w.Sender)
	if err != nil {
		return
	}
	return nil
}

func (w *PrepareContentBuilder) HexDump(prefix string, offsetFromStart membuffers.Offset) (err error) {
	if w == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	w._builder.Reset()
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "PrepareContent.SignedHeader", w.SignedHeader)
	if err != nil {
		return
	}
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "PrepareContent.Sender", w.Sender)
	if err != nil {
		return
	}
	return nil
}

func (w *PrepareContentBuilder) GetSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	return w._builder.GetSize()
}

func (w *PrepareContentBuilder) CalcRequiredSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	w.Write(nil)
	return w._builder.GetSize()
}

func (w *PrepareContentBuilder) Build() *PrepareContent {
	buf := make([]byte, w.CalcRequiredSize())
	if w.Write(buf) != nil {
		return nil
	}
	return PrepareContentReader(buf)
}

func PrepareContentBuilderFromRaw(raw []byte) *PrepareContentBuilder {
	return &PrepareContentBuilder{_overrideWithRawBuffer: raw}
}

/////////////////////////////////////////////////////////////////////////////
// message CommitContent

// reader

type CommitContent struct {
	// SignedHeader BlockRef
	// Sender SenderSignature
	// Share RandomSeedShare

	// internal
	// implements membuffers.Message
	_message membuffers.InternalMessage
}

func (x *CommitContent) String() string {
	if x == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{SignedHeader:%s,Sender:%s,Share:%s,}", x.StringSignedHeader(), x.StringSender(), x.StringShare())
}

var _CommitContent_Scheme = []membuffers.FieldType{membuffers.TypeMessage, membuffers.TypeMessage, membuffers.TypeMessage}
var _CommitContent_Unions = [][]membuffers.FieldType{}

func CommitContentReader(buf []byte) *CommitContent {
	x := &CommitContent{}
	x._message.Init(buf, membuffers.Offset(len(buf)), _CommitContent_Scheme, _CommitContent_Unions)
	return x
}

func (x *CommitContent) IsValid() bool {
	return x._message.IsValid()
}

func (x *CommitContent) Raw() []byte {
	return x._message.RawBuffer()
}

func (x *CommitContent) Equal(y *CommitContent) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return bytes.Equal(x.Raw(), y.Raw())
}

func (x *CommitContent) SignedHeader() *BlockRef {
	b, s := x._message.GetMessage(0)
	return BlockRefReader(b[:s])
}

func (x *CommitContent) RawSignedHeader() []byte {
	return x._message.RawBufferForField(0, 0)
}

func (x *CommitContent) RawSignedHeaderWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(0, 0)
}

func (x *CommitContent) StringSignedHeader() string {
	return x.SignedHeader().String()
}

func (x *CommitContent) Sender() *SenderSignature {
	b, s := x._message.GetMessage(1)
	return SenderSignatureReader(b[:s])
}

func (x *CommitContent) RawSender() []byte {
	return x._message.RawBufferForField(1, 0)
}

func (x *CommitContent) RawSenderWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(1, 0)
}

func (x *CommitContent) StringSender() string {
	return x.Sender().String()
}

func (x *CommitContent) Share() *RandomSeedShare {
	b, s := x._message.GetMessage(2)
	return RandomSeedShareReader(b[:s])
}

func (x *CommitContent) RawShare() []byte {
	return x._message.RawBufferForField(2, 0)
}

func (x *CommitContent) RawShareWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(2, 0)
}

func (x *CommitContent) StringShare() string {
	return x.Share().String()
}

// builder

type CommitContentBuilder struct {
	SignedHeader *BlockRefBuilder
	Sender       *SenderSignatureBuilder
	Share        *RandomSeedShareBuilder

	// internal
	// implements membuffers.Builder
	_builder               membuffers.InternalBuilder
	_overrideWithRawBuffer []byte
}

func (w *CommitContentBuilder) Write(buf []byte) (err error) {
	if w == nil {
		return
	}
	w._builder.NotifyBuildStart()
	defer w._builder.NotifyBuildEnd()
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	if w._overrideWithRawBuffer != nil {
		return w._builder.WriteOverrideWithRawBuffer(buf, w._overrideWithRawBuffer)
	}
	w._builder.Reset()
	err = w._builder.WriteMessage(buf, w.SignedHeader)
	if err != nil {
		return
	}
	err = w._builder.WriteMessage(buf, w.Sender)
	if err != nil {
		return
	}
	err = w._builder.WriteMessage(buf, w.Share)
	if err != nil {
		return
	}
	return nil
}

func (w *CommitContentBuilder) HexDump(prefix string, offsetFromStart membuffers.Offset) (err error) {
	if w == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	w._builder.Reset()
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "CommitContent.SignedHeader", w.SignedHeader)
	if err != nil {
		return
	}
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "CommitContent.Sender", w.Sender)
	if err != nil {
		return
	}
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "CommitContent.Share", w.Share)
	if err != nil {
		return
	}
	return nil
}

func (w *CommitContentBuilder) GetSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	return w._builder.GetSize()
}

func (w *CommitContentBuilder) CalcRequiredSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	w.Write(nil)
	return w._builder.GetSize()
}

func (w *CommitContentBuilder) Build() *CommitContent {
	buf := make([]byte, w.CalcRequiredSize())
	if w.Write(buf) != nil {
		return nil
	}
	return CommitContentReader(buf)
}

func CommitContentBuilderFromRaw(raw []byte) *CommitContentBuilder {
	return &CommitContentBuilder{_overrideWithRawBuffer: raw}
}

/////////////////////////////////////////////////////////////////////////////
// message ViewChangeMessageContent

// reader

type ViewChangeMessageContent struct {
	// SignedHeader ViewChangeHeader
	// Sender SenderSignature

	// internal
	// implements membuffers.Message
	_message membuffers.InternalMessage
}

func (x *ViewChangeMessageContent) String() string {
	if x == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{SignedHeader:%s,Sender:%s,}", x.StringSignedHeader(), x.StringSender())
}

var _ViewChangeMessageContent_Scheme = []membuffers.FieldType{membuffers.TypeMessage, membuffers.TypeMessage}
var _ViewChangeMessageContent_Unions = [][]membuffers.FieldType{}

func ViewChangeMessageContentReader(buf []byte) *ViewChangeMessageContent {
	x := &ViewChangeMessageContent{}
	x._message.Init(buf, membuffers.Offset(len(buf)), _ViewChangeMessageContent_Scheme, _ViewChangeMessageContent_Unions)
	return x
}

func (x *ViewChangeMessageContent) IsValid() bool {
	return x._message.IsValid()
}

func (x *ViewChangeMessageContent) Raw() []byte {
	return x._message.RawBuffer()
}

func (x *ViewChangeMessageContent) Equal(y *ViewChangeMessageContent) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return bytes.Equal(x.Raw(), y.Raw())
}

func (x *ViewChangeMessageContent) SignedHeader() *ViewChangeHeader {
	b, s := x._message.GetMessage(0)
	return ViewChangeHeaderReader(b[:s])
}

func (x *ViewChangeMessageContent) RawSignedHeader() []byte {
	return x._message.RawBufferForField(0, 0)
}

func (x *ViewChangeMessageContent) RawSignedHeaderWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(0, 0)
}

func (x *ViewChangeMessageContent) StringSignedHeader() string {
	return x.SignedHeader().String()
}

func (x *ViewChangeMessageContent) Sender() *SenderSignature {
	b, s := x._message.GetMessage(1)
	return SenderSignatureReader(b[:s])
}

func (x *ViewChangeMessageContent) RawSender() []byte {
	return x._message.RawBufferForField(1, 0)
}

func (x *ViewChangeMessageContent) RawSenderWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(1, 0)
}

func (x *ViewChangeMessageContent) StringSender() string {
	return x.Sender().String()
}

// builder

type ViewChangeMessageContentBuilder struct {
	SignedHeader *ViewChangeHeaderBuilder
	Sender       *SenderSignatureBuilder

	// internal
	// implements membuffers.Builder
	_builder               membuffers.InternalBuilder
	_overrideWithRawBuffer []byte
}

func (w *ViewChangeMessageContentBuilder) Write(buf []byte) (err error) {
	if w == nil {
		return
	}
	w._builder.NotifyBuildStart()
	defer w._builder.NotifyBuildEnd()
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	if w._overrideWithRawBuffer != nil {
		return w._builder.WriteOverrideWithRawBuffer(buf, w._overrideWithRawBuffer)
	}
	w._builder.Reset()
	err = w._builder.WriteMessage(buf, w.SignedHeader)
	if err != nil {
		return
	}
	err = w._builder.WriteMessage(buf, w.Sender)
	if err != nil {
		return
	}
	return nil
}

func (w *ViewChangeMessageContentBuilder) HexDump(prefix string, offsetFromStart membuffers.Offset) (err error) {
	if w == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	w._builder.Reset()
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "ViewChangeMessageContent.SignedHeader", w.SignedHeader)
	if err != nil {
		return
	}
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "ViewChangeMessageContent.Sender", w.Sender)
	if err != nil {
		return
	}
	return nil
}

func (w *ViewChangeMessageContentBuilder) GetSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	return w._builder.GetSize()
}

func (w *ViewChangeMessageContentBuilder) CalcRequiredSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	w.Write(nil)
	return w._builder.GetSize()
}

func (w *ViewChangeMessageContentBuilder) Build() *ViewChangeMessageContent {
	buf := make([]byte, w.CalcRequiredSize())
	if w.Write(buf) != nil {
		return nil
	}
	return ViewChangeMessageContentReader(buf)
}

func ViewChangeMessageContentBuilderFromRaw(raw []byte) *ViewChangeMessageContentBuilder {
	return &ViewChangeMessageContentBuilder{_overrideWithRawBuffer: raw}
}

/////////////////////////////////////////////////////////////////////////////
// message NewViewMessageContent

// reader

type NewViewMessageContent struct {
	// SignedHeader NewViewHeader
	// Sender SenderSignature
	// PreprepareMessageContent PreprepareContent

	// internal
	// implements membuffers.Message
	_message membuffers.InternalMessage
}

func (x *NewViewMessageContent) String() string {
	if x == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{SignedHeader:%s,Sender:%s,PreprepareMessageContent:%s,}", x.StringSignedHeader(), x.StringSender(), x.StringPreprepareMessageContent())
}

var _NewViewMessageContent_Scheme = []membuffers.FieldType{membuffers.TypeMessage, membuffers.TypeMessage, membuffers.TypeMessage}
var _NewViewMessageContent_Unions = [][]membuffers.FieldType{}

func NewViewMessageContentReader(buf []byte) *NewViewMessageContent {
	x := &NewViewMessageContent{}
	x._message.Init(buf, membuffers.Offset(len(buf)), _NewViewMessageContent_Scheme, _NewViewMessageContent_Unions)
	return x
}

func (x *NewViewMessageContent) IsValid() bool {
	return x._message.IsValid()
}

func (x *NewViewMessageContent) Raw() []byte {
	return x._message.RawBuffer()
}

func (x *NewViewMessageContent) Equal(y *NewViewMessageContent) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return bytes.Equal(x.Raw(), y.Raw())
}

func (x *NewViewMessageContent) SignedHeader() *NewViewHeader {
	b, s := x._message.GetMessage(0)
	return NewViewHeaderReader(b[:s])
}

func (x *NewViewMessageContent) RawSignedHeader() []byte {
	return x._message.RawBufferForField(0, 0)
}

func (x *NewViewMessageContent) RawSignedHeaderWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(0, 0)
}

func (x *NewViewMessageContent) StringSignedHeader() string {
	return x.SignedHeader().String()
}

func (x *NewViewMessageContent) Sender() *SenderSignature {
	b, s := x._message.GetMessage(1)
	return SenderSignatureReader(b[:s])
}

func (x *NewViewMessageContent) RawSender() []byte {
	return x._message.RawBufferForField(1, 0)
}

func (x *NewViewMessageContent) RawSenderWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(1, 0)
}

func (x *NewViewMessageContent) StringSender() string {
	return x.Sender().String()
}

func (x *NewViewMessageContent) PreprepareMessageContent() *PreprepareContent {
	b, s := x._message.GetMessage(2)
	return PreprepareContentReader(b[:s])
}

func (x *NewViewMessageContent) RawPreprepareMessageContent() []byte {
	return x._message.RawBufferForField(2, 0)
}

func (x *NewViewMessageContent) RawPreprepareMessageContentWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(2, 0)
}

func (x *NewViewMessageContent) StringPreprepareMessageContent() string {
	return x.PreprepareMessageContent().String()
}

// builder

type NewViewMessageContentBuilder struct {
	SignedHeader             *NewViewHeaderBuilder
	Sender                   *SenderSignatureBuilder
	PreprepareMessageContent *PreprepareContentBuilder

	// internal
	// implements membuffers.Builder
	_builder               membuffers.InternalBuilder
	_overrideWithRawBuffer []byte
}

func (w *NewViewMessageContentBuilder) Write(buf []byte) (err error) {
	if w == nil {
		return
	}
	w._builder.NotifyBuildStart()
	defer w._builder.NotifyBuildEnd()
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	if w._overrideWithRawBuffer != nil {
		return w._builder.WriteOverrideWithRawBuffer(buf, w._overrideWithRawBuffer)
	}
	w._builder.Reset()
	err = w._builder.WriteMessage(buf, w.SignedHeader)
	if err != nil {
		return
	}
	err = w._builder.WriteMessage(buf, w.Sender)
	if err != nil {
		return
	}
	err = w._builder.WriteMessage(buf, w.PreprepareMessageContent)
	if err != nil {
		return
	}
	return nil
}

func (w *NewViewMessageContentBuilder) HexDump(prefix string, offsetFromStart membuffers.Offset) (err error) {
	if w == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	w._builder.Reset()
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "NewViewMessageContent.SignedHeader", w.SignedHeader)
	if err != nil {
		return
	}
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "NewViewMessageContent.Sender", w.Sender)
	if err != nil {
		return
	}
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "NewViewMessageContent.PreprepareMessageContent", w.PreprepareMessageContent)
	if err != nil {
		return
	}
	return nil
}

func (w *NewViewMessageContentBuilder) GetSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	return w._builder.GetSize()
}

func (w *NewViewMessageContentBuilder) CalcRequiredSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	w.Write(nil)
	return w._builder.GetSize()
}

func (w *NewViewMessageContentBuilder) Build() *NewViewMessageContent {
	buf := make([]byte, w.CalcRequiredSize())
	if w.Write(buf) != nil {
		return nil
	}
	return NewViewMessageContentReader(buf)
}

func NewViewMessageContentBuilderFromRaw(raw []byte) *NewViewMessageContentBuilder {
	return &NewViewMessageContentBuilder{_overrideWithRawBuffer: raw}
}

/////////////////////////////////////////////////////////////////////////////
// message SenderSignature

// reader

type SenderSignature struct {
	// SenderPublicKey primitives.Ed25519PublicKey
	// Signature primitives.Ed25519Sig

	// internal
	// implements membuffers.Message
	_message membuffers.InternalMessage
}

func (x *SenderSignature) String() string {
	if x == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{SenderPublicKey:%s,Signature:%s,}", x.StringSenderPublicKey(), x.StringSignature())
}

var _SenderSignature_Scheme = []membuffers.FieldType{membuffers.TypeBytes, membuffers.TypeBytes}
var _SenderSignature_Unions = [][]membuffers.FieldType{}

func SenderSignatureReader(buf []byte) *SenderSignature {
	x := &SenderSignature{}
	x._message.Init(buf, membuffers.Offset(len(buf)), _SenderSignature_Scheme, _SenderSignature_Unions)
	return x
}

func (x *SenderSignature) IsValid() bool {
	return x._message.IsValid()
}

func (x *SenderSignature) Raw() []byte {
	return x._message.RawBuffer()
}

func (x *SenderSignature) Equal(y *SenderSignature) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return bytes.Equal(x.Raw(), y.Raw())
}

func (x *SenderSignature) SenderPublicKey() primitives.Ed25519PublicKey {
	return primitives.Ed25519PublicKey(x._message.GetBytes(0))
}

func (x *SenderSignature) RawSenderPublicKey() []byte {
	return x._message.RawBufferForField(0, 0)
}

func (x *SenderSignature) MutateSenderPublicKey(v primitives.Ed25519PublicKey) error {
	return x._message.SetBytes(0, []byte(v))
}

func (x *SenderSignature) StringSenderPublicKey() string {
	return fmt.Sprintf("%s", x.SenderPublicKey())
}

func (x *SenderSignature) Signature() primitives.Ed25519Sig {
	return primitives.Ed25519Sig(x._message.GetBytes(1))
}

func (x *SenderSignature) RawSignature() []byte {
	return x._message.RawBufferForField(1, 0)
}

func (x *SenderSignature) MutateSignature(v primitives.Ed25519Sig) error {
	return x._message.SetBytes(1, []byte(v))
}

func (x *SenderSignature) StringSignature() string {
	return fmt.Sprintf("%s", x.Signature())
}

// builder

type SenderSignatureBuilder struct {
	SenderPublicKey primitives.Ed25519PublicKey
	Signature       primitives.Ed25519Sig

	// internal
	// implements membuffers.Builder
	_builder               membuffers.InternalBuilder
	_overrideWithRawBuffer []byte
}

func (w *SenderSignatureBuilder) Write(buf []byte) (err error) {
	if w == nil {
		return
	}
	w._builder.NotifyBuildStart()
	defer w._builder.NotifyBuildEnd()
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	if w._overrideWithRawBuffer != nil {
		return w._builder.WriteOverrideWithRawBuffer(buf, w._overrideWithRawBuffer)
	}
	w._builder.Reset()
	w._builder.WriteBytes(buf, []byte(w.SenderPublicKey))
	w._builder.WriteBytes(buf, []byte(w.Signature))
	return nil
}

func (w *SenderSignatureBuilder) HexDump(prefix string, offsetFromStart membuffers.Offset) (err error) {
	if w == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	w._builder.Reset()
	w._builder.HexDumpBytes(prefix, offsetFromStart, "SenderSignature.SenderPublicKey", []byte(w.SenderPublicKey))
	w._builder.HexDumpBytes(prefix, offsetFromStart, "SenderSignature.Signature", []byte(w.Signature))
	return nil
}

func (w *SenderSignatureBuilder) GetSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	return w._builder.GetSize()
}

func (w *SenderSignatureBuilder) CalcRequiredSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	w.Write(nil)
	return w._builder.GetSize()
}

func (w *SenderSignatureBuilder) Build() *SenderSignature {
	buf := make([]byte, w.CalcRequiredSize())
	if w.Write(buf) != nil {
		return nil
	}
	return SenderSignatureReader(buf)
}

func SenderSignatureBuilderFromRaw(raw []byte) *SenderSignatureBuilder {
	return &SenderSignatureBuilder{_overrideWithRawBuffer: raw}
}

/////////////////////////////////////////////////////////////////////////////
// message RandomSeedShare

// reader

type RandomSeedShare struct {
	// SenderPublicKey primitives.Bls1PublicKey
	// Signature primitives.Bls1Sig

	// internal
	// implements membuffers.Message
	_message membuffers.InternalMessage
}

func (x *RandomSeedShare) String() string {
	if x == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{SenderPublicKey:%s,Signature:%s,}", x.StringSenderPublicKey(), x.StringSignature())
}

var _RandomSeedShare_Scheme = []membuffers.FieldType{membuffers.TypeBytes, membuffers.TypeBytes}
var _RandomSeedShare_Unions = [][]membuffers.FieldType{}

func RandomSeedShareReader(buf []byte) *RandomSeedShare {
	x := &RandomSeedShare{}
	x._message.Init(buf, membuffers.Offset(len(buf)), _RandomSeedShare_Scheme, _RandomSeedShare_Unions)
	return x
}

func (x *RandomSeedShare) IsValid() bool {
	return x._message.IsValid()
}

func (x *RandomSeedShare) Raw() []byte {
	return x._message.RawBuffer()
}

func (x *RandomSeedShare) Equal(y *RandomSeedShare) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return bytes.Equal(x.Raw(), y.Raw())
}

func (x *RandomSeedShare) SenderPublicKey() primitives.Bls1PublicKey {
	return primitives.Bls1PublicKey(x._message.GetBytes(0))
}

func (x *RandomSeedShare) RawSenderPublicKey() []byte {
	return x._message.RawBufferForField(0, 0)
}

func (x *RandomSeedShare) MutateSenderPublicKey(v primitives.Bls1PublicKey) error {
	return x._message.SetBytes(0, []byte(v))
}

func (x *RandomSeedShare) StringSenderPublicKey() string {
	return fmt.Sprintf("%s", x.SenderPublicKey())
}

func (x *RandomSeedShare) Signature() primitives.Bls1Sig {
	return primitives.Bls1Sig(x._message.GetBytes(1))
}

func (x *RandomSeedShare) RawSignature() []byte {
	return x._message.RawBufferForField(1, 0)
}

func (x *RandomSeedShare) MutateSignature(v primitives.Bls1Sig) error {
	return x._message.SetBytes(1, []byte(v))
}

func (x *RandomSeedShare) StringSignature() string {
	return fmt.Sprintf("%s", x.Signature())
}

// builder

type RandomSeedShareBuilder struct {
	SenderPublicKey primitives.Bls1PublicKey
	Signature       primitives.Bls1Sig

	// internal
	// implements membuffers.Builder
	_builder               membuffers.InternalBuilder
	_overrideWithRawBuffer []byte
}

func (w *RandomSeedShareBuilder) Write(buf []byte) (err error) {
	if w == nil {
		return
	}
	w._builder.NotifyBuildStart()
	defer w._builder.NotifyBuildEnd()
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	if w._overrideWithRawBuffer != nil {
		return w._builder.WriteOverrideWithRawBuffer(buf, w._overrideWithRawBuffer)
	}
	w._builder.Reset()
	w._builder.WriteBytes(buf, []byte(w.SenderPublicKey))
	w._builder.WriteBytes(buf, []byte(w.Signature))
	return nil
}

func (w *RandomSeedShareBuilder) HexDump(prefix string, offsetFromStart membuffers.Offset) (err error) {
	if w == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	w._builder.Reset()
	w._builder.HexDumpBytes(prefix, offsetFromStart, "RandomSeedShare.SenderPublicKey", []byte(w.SenderPublicKey))
	w._builder.HexDumpBytes(prefix, offsetFromStart, "RandomSeedShare.Signature", []byte(w.Signature))
	return nil
}

func (w *RandomSeedShareBuilder) GetSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	return w._builder.GetSize()
}

func (w *RandomSeedShareBuilder) CalcRequiredSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	w.Write(nil)
	return w._builder.GetSize()
}

func (w *RandomSeedShareBuilder) Build() *RandomSeedShare {
	buf := make([]byte, w.CalcRequiredSize())
	if w.Write(buf) != nil {
		return nil
	}
	return RandomSeedShareReader(buf)
}

func RandomSeedShareBuilderFromRaw(raw []byte) *RandomSeedShareBuilder {
	return &RandomSeedShareBuilder{_overrideWithRawBuffer: raw}
}

/////////////////////////////////////////////////////////////////////////////
// message BlockRef

// reader

type BlockRef struct {
	// MessageType MessageType
	// BlockHeight primitives.BlockHeight
	// View primitives.View
	// BlockHash primitives.Uint256

	// internal
	// implements membuffers.Message
	_message membuffers.InternalMessage
}

func (x *BlockRef) String() string {
	if x == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{MessageType:%s,BlockHeight:%s,View:%s,BlockHash:%s,}", x.StringMessageType(), x.StringBlockHeight(), x.StringView(), x.StringBlockHash())
}

var _BlockRef_Scheme = []membuffers.FieldType{membuffers.TypeUint16, membuffers.TypeUint64, membuffers.TypeUint64, membuffers.TypeBytes}
var _BlockRef_Unions = [][]membuffers.FieldType{}

func BlockRefReader(buf []byte) *BlockRef {
	x := &BlockRef{}
	x._message.Init(buf, membuffers.Offset(len(buf)), _BlockRef_Scheme, _BlockRef_Unions)
	return x
}

func (x *BlockRef) IsValid() bool {
	return x._message.IsValid()
}

func (x *BlockRef) Raw() []byte {
	return x._message.RawBuffer()
}

func (x *BlockRef) Equal(y *BlockRef) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return bytes.Equal(x.Raw(), y.Raw())
}

func (x *BlockRef) MessageType() MessageType {
	return MessageType(x._message.GetUint16(0))
}

func (x *BlockRef) RawMessageType() []byte {
	return x._message.RawBufferForField(0, 0)
}

func (x *BlockRef) MutateMessageType(v MessageType) error {
	return x._message.SetUint16(0, uint16(v))
}

func (x *BlockRef) StringMessageType() string {
	return x.MessageType().String()
}

func (x *BlockRef) BlockHeight() primitives.BlockHeight {
	return primitives.BlockHeight(x._message.GetUint64(1))
}

func (x *BlockRef) RawBlockHeight() []byte {
	return x._message.RawBufferForField(1, 0)
}

func (x *BlockRef) MutateBlockHeight(v primitives.BlockHeight) error {
	return x._message.SetUint64(1, uint64(v))
}

func (x *BlockRef) StringBlockHeight() string {
	return fmt.Sprintf("%s", x.BlockHeight())
}

func (x *BlockRef) View() primitives.View {
	return primitives.View(x._message.GetUint64(2))
}

func (x *BlockRef) RawView() []byte {
	return x._message.RawBufferForField(2, 0)
}

func (x *BlockRef) MutateView(v primitives.View) error {
	return x._message.SetUint64(2, uint64(v))
}

func (x *BlockRef) StringView() string {
	return fmt.Sprintf("%s", x.View())
}

func (x *BlockRef) BlockHash() primitives.Uint256 {
	return primitives.Uint256(x._message.GetBytes(3))
}

func (x *BlockRef) RawBlockHash() []byte {
	return x._message.RawBufferForField(3, 0)
}

func (x *BlockRef) MutateBlockHash(v primitives.Uint256) error {
	return x._message.SetBytes(3, []byte(v))
}

func (x *BlockRef) StringBlockHash() string {
	return fmt.Sprintf("%s", x.BlockHash())
}

// builder

type BlockRefBuilder struct {
	MessageType MessageType
	BlockHeight primitives.BlockHeight
	View        primitives.View
	BlockHash   primitives.Uint256

	// internal
	// implements membuffers.Builder
	_builder               membuffers.InternalBuilder
	_overrideWithRawBuffer []byte
}

func (w *BlockRefBuilder) Write(buf []byte) (err error) {
	if w == nil {
		return
	}
	w._builder.NotifyBuildStart()
	defer w._builder.NotifyBuildEnd()
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	if w._overrideWithRawBuffer != nil {
		return w._builder.WriteOverrideWithRawBuffer(buf, w._overrideWithRawBuffer)
	}
	w._builder.Reset()
	w._builder.WriteUint16(buf, uint16(w.MessageType))
	w._builder.WriteUint64(buf, uint64(w.BlockHeight))
	w._builder.WriteUint64(buf, uint64(w.View))
	w._builder.WriteBytes(buf, []byte(w.BlockHash))
	return nil
}

func (w *BlockRefBuilder) HexDump(prefix string, offsetFromStart membuffers.Offset) (err error) {
	if w == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	w._builder.Reset()
	w._builder.HexDumpUint16(prefix, offsetFromStart, "BlockRef.MessageType", uint16(w.MessageType))
	w._builder.HexDumpUint64(prefix, offsetFromStart, "BlockRef.BlockHeight", uint64(w.BlockHeight))
	w._builder.HexDumpUint64(prefix, offsetFromStart, "BlockRef.View", uint64(w.View))
	w._builder.HexDumpBytes(prefix, offsetFromStart, "BlockRef.BlockHash", []byte(w.BlockHash))
	return nil
}

func (w *BlockRefBuilder) GetSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	return w._builder.GetSize()
}

func (w *BlockRefBuilder) CalcRequiredSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	w.Write(nil)
	return w._builder.GetSize()
}

func (w *BlockRefBuilder) Build() *BlockRef {
	buf := make([]byte, w.CalcRequiredSize())
	if w.Write(buf) != nil {
		return nil
	}
	return BlockRefReader(buf)
}

func BlockRefBuilderFromRaw(raw []byte) *BlockRefBuilder {
	return &BlockRefBuilder{_overrideWithRawBuffer: raw}
}

/////////////////////////////////////////////////////////////////////////////
// message ViewChangeHeader

// reader

type ViewChangeHeader struct {
	// MessageType MessageType
	// BlockHeight primitives.BlockHeight
	// View primitives.View
	// PreparedProof PreparedProof

	// internal
	// implements membuffers.Message
	_message membuffers.InternalMessage
}

func (x *ViewChangeHeader) String() string {
	if x == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{MessageType:%s,BlockHeight:%s,View:%s,PreparedProof:%s,}", x.StringMessageType(), x.StringBlockHeight(), x.StringView(), x.StringPreparedProof())
}

var _ViewChangeHeader_Scheme = []membuffers.FieldType{membuffers.TypeUint16, membuffers.TypeUint64, membuffers.TypeUint64, membuffers.TypeMessage}
var _ViewChangeHeader_Unions = [][]membuffers.FieldType{}

func ViewChangeHeaderReader(buf []byte) *ViewChangeHeader {
	x := &ViewChangeHeader{}
	x._message.Init(buf, membuffers.Offset(len(buf)), _ViewChangeHeader_Scheme, _ViewChangeHeader_Unions)
	return x
}

func (x *ViewChangeHeader) IsValid() bool {
	return x._message.IsValid()
}

func (x *ViewChangeHeader) Raw() []byte {
	return x._message.RawBuffer()
}

func (x *ViewChangeHeader) Equal(y *ViewChangeHeader) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return bytes.Equal(x.Raw(), y.Raw())
}

func (x *ViewChangeHeader) MessageType() MessageType {
	return MessageType(x._message.GetUint16(0))
}

func (x *ViewChangeHeader) RawMessageType() []byte {
	return x._message.RawBufferForField(0, 0)
}

func (x *ViewChangeHeader) MutateMessageType(v MessageType) error {
	return x._message.SetUint16(0, uint16(v))
}

func (x *ViewChangeHeader) StringMessageType() string {
	return x.MessageType().String()
}

func (x *ViewChangeHeader) BlockHeight() primitives.BlockHeight {
	return primitives.BlockHeight(x._message.GetUint64(1))
}

func (x *ViewChangeHeader) RawBlockHeight() []byte {
	return x._message.RawBufferForField(1, 0)
}

func (x *ViewChangeHeader) MutateBlockHeight(v primitives.BlockHeight) error {
	return x._message.SetUint64(1, uint64(v))
}

func (x *ViewChangeHeader) StringBlockHeight() string {
	return fmt.Sprintf("%s", x.BlockHeight())
}

func (x *ViewChangeHeader) View() primitives.View {
	return primitives.View(x._message.GetUint64(2))
}

func (x *ViewChangeHeader) RawView() []byte {
	return x._message.RawBufferForField(2, 0)
}

func (x *ViewChangeHeader) MutateView(v primitives.View) error {
	return x._message.SetUint64(2, uint64(v))
}

func (x *ViewChangeHeader) StringView() string {
	return fmt.Sprintf("%s", x.View())
}

func (x *ViewChangeHeader) PreparedProof() *PreparedProof {
	b, s := x._message.GetMessage(3)
	return PreparedProofReader(b[:s])
}

func (x *ViewChangeHeader) RawPreparedProof() []byte {
	return x._message.RawBufferForField(3, 0)
}

func (x *ViewChangeHeader) RawPreparedProofWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(3, 0)
}

func (x *ViewChangeHeader) StringPreparedProof() string {
	return x.PreparedProof().String()
}

// builder

type ViewChangeHeaderBuilder struct {
	MessageType   MessageType
	BlockHeight   primitives.BlockHeight
	View          primitives.View
	PreparedProof *PreparedProofBuilder

	// internal
	// implements membuffers.Builder
	_builder               membuffers.InternalBuilder
	_overrideWithRawBuffer []byte
}

func (w *ViewChangeHeaderBuilder) Write(buf []byte) (err error) {
	if w == nil {
		return
	}
	w._builder.NotifyBuildStart()
	defer w._builder.NotifyBuildEnd()
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	if w._overrideWithRawBuffer != nil {
		return w._builder.WriteOverrideWithRawBuffer(buf, w._overrideWithRawBuffer)
	}
	w._builder.Reset()
	w._builder.WriteUint16(buf, uint16(w.MessageType))
	w._builder.WriteUint64(buf, uint64(w.BlockHeight))
	w._builder.WriteUint64(buf, uint64(w.View))
	err = w._builder.WriteMessage(buf, w.PreparedProof)
	if err != nil {
		return
	}
	return nil
}

func (w *ViewChangeHeaderBuilder) HexDump(prefix string, offsetFromStart membuffers.Offset) (err error) {
	if w == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	w._builder.Reset()
	w._builder.HexDumpUint16(prefix, offsetFromStart, "ViewChangeHeader.MessageType", uint16(w.MessageType))
	w._builder.HexDumpUint64(prefix, offsetFromStart, "ViewChangeHeader.BlockHeight", uint64(w.BlockHeight))
	w._builder.HexDumpUint64(prefix, offsetFromStart, "ViewChangeHeader.View", uint64(w.View))
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "ViewChangeHeader.PreparedProof", w.PreparedProof)
	if err != nil {
		return
	}
	return nil
}

func (w *ViewChangeHeaderBuilder) GetSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	return w._builder.GetSize()
}

func (w *ViewChangeHeaderBuilder) CalcRequiredSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	w.Write(nil)
	return w._builder.GetSize()
}

func (w *ViewChangeHeaderBuilder) Build() *ViewChangeHeader {
	buf := make([]byte, w.CalcRequiredSize())
	if w.Write(buf) != nil {
		return nil
	}
	return ViewChangeHeaderReader(buf)
}

func ViewChangeHeaderBuilderFromRaw(raw []byte) *ViewChangeHeaderBuilder {
	return &ViewChangeHeaderBuilder{_overrideWithRawBuffer: raw}
}

/////////////////////////////////////////////////////////////////////////////
// message PreparedProof

// reader

type PreparedProof struct {
	// PreprepareBlockRef BlockRef
	// PreprepareSender SenderSignature
	// PrepareBlockRef BlockRef
	// PrepareSenders []SenderSignature

	// internal
	// implements membuffers.Message
	_message membuffers.InternalMessage
}

func (x *PreparedProof) String() string {
	if x == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{PreprepareBlockRef:%s,PreprepareSender:%s,PrepareBlockRef:%s,PrepareSenders:%s,}", x.StringPreprepareBlockRef(), x.StringPreprepareSender(), x.StringPrepareBlockRef(), x.StringPrepareSenders())
}

var _PreparedProof_Scheme = []membuffers.FieldType{membuffers.TypeMessage, membuffers.TypeMessage, membuffers.TypeMessage, membuffers.TypeMessageArray}
var _PreparedProof_Unions = [][]membuffers.FieldType{}

func PreparedProofReader(buf []byte) *PreparedProof {
	x := &PreparedProof{}
	x._message.Init(buf, membuffers.Offset(len(buf)), _PreparedProof_Scheme, _PreparedProof_Unions)
	return x
}

func (x *PreparedProof) IsValid() bool {
	return x._message.IsValid()
}

func (x *PreparedProof) Raw() []byte {
	return x._message.RawBuffer()
}

func (x *PreparedProof) Equal(y *PreparedProof) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return bytes.Equal(x.Raw(), y.Raw())
}

func (x *PreparedProof) PreprepareBlockRef() *BlockRef {
	b, s := x._message.GetMessage(0)
	return BlockRefReader(b[:s])
}

func (x *PreparedProof) RawPreprepareBlockRef() []byte {
	return x._message.RawBufferForField(0, 0)
}

func (x *PreparedProof) RawPreprepareBlockRefWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(0, 0)
}

func (x *PreparedProof) StringPreprepareBlockRef() string {
	return x.PreprepareBlockRef().String()
}

func (x *PreparedProof) PreprepareSender() *SenderSignature {
	b, s := x._message.GetMessage(1)
	return SenderSignatureReader(b[:s])
}

func (x *PreparedProof) RawPreprepareSender() []byte {
	return x._message.RawBufferForField(1, 0)
}

func (x *PreparedProof) RawPreprepareSenderWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(1, 0)
}

func (x *PreparedProof) StringPreprepareSender() string {
	return x.PreprepareSender().String()
}

func (x *PreparedProof) PrepareBlockRef() *BlockRef {
	b, s := x._message.GetMessage(2)
	return BlockRefReader(b[:s])
}

func (x *PreparedProof) RawPrepareBlockRef() []byte {
	return x._message.RawBufferForField(2, 0)
}

func (x *PreparedProof) RawPrepareBlockRefWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(2, 0)
}

func (x *PreparedProof) StringPrepareBlockRef() string {
	return x.PrepareBlockRef().String()
}

func (x *PreparedProof) PrepareSendersIterator() *PreparedProofPrepareSendersIterator {
	return &PreparedProofPrepareSendersIterator{iterator: x._message.GetMessageArrayIterator(3)}
}

type PreparedProofPrepareSendersIterator struct {
	iterator *membuffers.Iterator
}

func (i *PreparedProofPrepareSendersIterator) HasNext() bool {
	return i.iterator.HasNext()
}

func (i *PreparedProofPrepareSendersIterator) NextPrepareSenders() *SenderSignature {
	b, s := i.iterator.NextMessage()
	return SenderSignatureReader(b[:s])
}

func (x *PreparedProof) RawPrepareSendersArray() []byte {
	return x._message.RawBufferForField(3, 0)
}

func (x *PreparedProof) RawPrepareSendersArrayWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(3, 0)
}

func (x *PreparedProof) StringPrepareSenders() (res string) {
	res = "["
	for i := x.PrepareSendersIterator(); i.HasNext(); {
		res += i.NextPrepareSenders().String() + ","
	}
	res += "]"
	return
}

// builder

type PreparedProofBuilder struct {
	PreprepareBlockRef *BlockRefBuilder
	PreprepareSender   *SenderSignatureBuilder
	PrepareBlockRef    *BlockRefBuilder
	PrepareSenders     []*SenderSignatureBuilder

	// internal
	// implements membuffers.Builder
	_builder               membuffers.InternalBuilder
	_overrideWithRawBuffer []byte
}

func (w *PreparedProofBuilder) arrayOfPrepareSenders() []membuffers.MessageWriter {
	res := make([]membuffers.MessageWriter, len(w.PrepareSenders))
	for i, v := range w.PrepareSenders {
		res[i] = v
	}
	return res
}

func (w *PreparedProofBuilder) Write(buf []byte) (err error) {
	if w == nil {
		return
	}
	w._builder.NotifyBuildStart()
	defer w._builder.NotifyBuildEnd()
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	if w._overrideWithRawBuffer != nil {
		return w._builder.WriteOverrideWithRawBuffer(buf, w._overrideWithRawBuffer)
	}
	w._builder.Reset()
	err = w._builder.WriteMessage(buf, w.PreprepareBlockRef)
	if err != nil {
		return
	}
	err = w._builder.WriteMessage(buf, w.PreprepareSender)
	if err != nil {
		return
	}
	err = w._builder.WriteMessage(buf, w.PrepareBlockRef)
	if err != nil {
		return
	}
	err = w._builder.WriteMessageArray(buf, w.arrayOfPrepareSenders())
	if err != nil {
		return
	}
	return nil
}

func (w *PreparedProofBuilder) HexDump(prefix string, offsetFromStart membuffers.Offset) (err error) {
	if w == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	w._builder.Reset()
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "PreparedProof.PreprepareBlockRef", w.PreprepareBlockRef)
	if err != nil {
		return
	}
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "PreparedProof.PreprepareSender", w.PreprepareSender)
	if err != nil {
		return
	}
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "PreparedProof.PrepareBlockRef", w.PrepareBlockRef)
	if err != nil {
		return
	}
	err = w._builder.HexDumpMessageArray(prefix, offsetFromStart, "PreparedProof.PrepareSenders", w.arrayOfPrepareSenders())
	if err != nil {
		return
	}
	return nil
}

func (w *PreparedProofBuilder) GetSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	return w._builder.GetSize()
}

func (w *PreparedProofBuilder) CalcRequiredSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	w.Write(nil)
	return w._builder.GetSize()
}

func (w *PreparedProofBuilder) Build() *PreparedProof {
	buf := make([]byte, w.CalcRequiredSize())
	if w.Write(buf) != nil {
		return nil
	}
	return PreparedProofReader(buf)
}

func PreparedProofBuilderFromRaw(raw []byte) *PreparedProofBuilder {
	return &PreparedProofBuilder{_overrideWithRawBuffer: raw}
}

/////////////////////////////////////////////////////////////////////////////
// message NewViewHeader

// reader

type NewViewHeader struct {
	// MessageType MessageType
	// BlockHeight primitives.BlockHeight
	// View primitives.View
	// ViewChangeConfirmations []ViewChangeMessageContent

	// internal
	// implements membuffers.Message
	_message membuffers.InternalMessage
}

func (x *NewViewHeader) String() string {
	if x == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{MessageType:%s,BlockHeight:%s,View:%s,ViewChangeConfirmations:%s,}", x.StringMessageType(), x.StringBlockHeight(), x.StringView(), x.StringViewChangeConfirmations())
}

var _NewViewHeader_Scheme = []membuffers.FieldType{membuffers.TypeUint16, membuffers.TypeUint64, membuffers.TypeUint64, membuffers.TypeMessageArray}
var _NewViewHeader_Unions = [][]membuffers.FieldType{}

func NewViewHeaderReader(buf []byte) *NewViewHeader {
	x := &NewViewHeader{}
	x._message.Init(buf, membuffers.Offset(len(buf)), _NewViewHeader_Scheme, _NewViewHeader_Unions)
	return x
}

func (x *NewViewHeader) IsValid() bool {
	return x._message.IsValid()
}

func (x *NewViewHeader) Raw() []byte {
	return x._message.RawBuffer()
}

func (x *NewViewHeader) Equal(y *NewViewHeader) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return bytes.Equal(x.Raw(), y.Raw())
}

func (x *NewViewHeader) MessageType() MessageType {
	return MessageType(x._message.GetUint16(0))
}

func (x *NewViewHeader) RawMessageType() []byte {
	return x._message.RawBufferForField(0, 0)
}

func (x *NewViewHeader) MutateMessageType(v MessageType) error {
	return x._message.SetUint16(0, uint16(v))
}

func (x *NewViewHeader) StringMessageType() string {
	return x.MessageType().String()
}

func (x *NewViewHeader) BlockHeight() primitives.BlockHeight {
	return primitives.BlockHeight(x._message.GetUint64(1))
}

func (x *NewViewHeader) RawBlockHeight() []byte {
	return x._message.RawBufferForField(1, 0)
}

func (x *NewViewHeader) MutateBlockHeight(v primitives.BlockHeight) error {
	return x._message.SetUint64(1, uint64(v))
}

func (x *NewViewHeader) StringBlockHeight() string {
	return fmt.Sprintf("%s", x.BlockHeight())
}

func (x *NewViewHeader) View() primitives.View {
	return primitives.View(x._message.GetUint64(2))
}

func (x *NewViewHeader) RawView() []byte {
	return x._message.RawBufferForField(2, 0)
}

func (x *NewViewHeader) MutateView(v primitives.View) error {
	return x._message.SetUint64(2, uint64(v))
}

func (x *NewViewHeader) StringView() string {
	return fmt.Sprintf("%s", x.View())
}

func (x *NewViewHeader) ViewChangeConfirmationsIterator() *NewViewHeaderViewChangeConfirmationsIterator {
	return &NewViewHeaderViewChangeConfirmationsIterator{iterator: x._message.GetMessageArrayIterator(3)}
}

type NewViewHeaderViewChangeConfirmationsIterator struct {
	iterator *membuffers.Iterator
}

func (i *NewViewHeaderViewChangeConfirmationsIterator) HasNext() bool {
	return i.iterator.HasNext()
}

func (i *NewViewHeaderViewChangeConfirmationsIterator) NextViewChangeConfirmations() *ViewChangeMessageContent {
	b, s := i.iterator.NextMessage()
	return ViewChangeMessageContentReader(b[:s])
}

func (x *NewViewHeader) RawViewChangeConfirmationsArray() []byte {
	return x._message.RawBufferForField(3, 0)
}

func (x *NewViewHeader) RawViewChangeConfirmationsArrayWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(3, 0)
}

func (x *NewViewHeader) StringViewChangeConfirmations() (res string) {
	res = "["
	for i := x.ViewChangeConfirmationsIterator(); i.HasNext(); {
		res += i.NextViewChangeConfirmations().String() + ","
	}
	res += "]"
	return
}

// builder

type NewViewHeaderBuilder struct {
	MessageType             MessageType
	BlockHeight             primitives.BlockHeight
	View                    primitives.View
	ViewChangeConfirmations []*ViewChangeMessageContentBuilder

	// internal
	// implements membuffers.Builder
	_builder               membuffers.InternalBuilder
	_overrideWithRawBuffer []byte
}

func (w *NewViewHeaderBuilder) arrayOfViewChangeConfirmations() []membuffers.MessageWriter {
	res := make([]membuffers.MessageWriter, len(w.ViewChangeConfirmations))
	for i, v := range w.ViewChangeConfirmations {
		res[i] = v
	}
	return res
}

func (w *NewViewHeaderBuilder) Write(buf []byte) (err error) {
	if w == nil {
		return
	}
	w._builder.NotifyBuildStart()
	defer w._builder.NotifyBuildEnd()
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	if w._overrideWithRawBuffer != nil {
		return w._builder.WriteOverrideWithRawBuffer(buf, w._overrideWithRawBuffer)
	}
	w._builder.Reset()
	w._builder.WriteUint16(buf, uint16(w.MessageType))
	w._builder.WriteUint64(buf, uint64(w.BlockHeight))
	w._builder.WriteUint64(buf, uint64(w.View))
	err = w._builder.WriteMessageArray(buf, w.arrayOfViewChangeConfirmations())
	if err != nil {
		return
	}
	return nil
}

func (w *NewViewHeaderBuilder) HexDump(prefix string, offsetFromStart membuffers.Offset) (err error) {
	if w == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	w._builder.Reset()
	w._builder.HexDumpUint16(prefix, offsetFromStart, "NewViewHeader.MessageType", uint16(w.MessageType))
	w._builder.HexDumpUint64(prefix, offsetFromStart, "NewViewHeader.BlockHeight", uint64(w.BlockHeight))
	w._builder.HexDumpUint64(prefix, offsetFromStart, "NewViewHeader.View", uint64(w.View))
	err = w._builder.HexDumpMessageArray(prefix, offsetFromStart, "NewViewHeader.ViewChangeConfirmations", w.arrayOfViewChangeConfirmations())
	if err != nil {
		return
	}
	return nil
}

func (w *NewViewHeaderBuilder) GetSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	return w._builder.GetSize()
}

func (w *NewViewHeaderBuilder) CalcRequiredSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	w.Write(nil)
	return w._builder.GetSize()
}

func (w *NewViewHeaderBuilder) Build() *NewViewHeader {
	buf := make([]byte, w.CalcRequiredSize())
	if w.Write(buf) != nil {
		return nil
	}
	return NewViewHeaderReader(buf)
}

func NewViewHeaderBuilderFromRaw(raw []byte) *NewViewHeaderBuilder {
	return &NewViewHeaderBuilder{_overrideWithRawBuffer: raw}
}

/////////////////////////////////////////////////////////////////////////////
// message BlockProof

// reader

type BlockProof struct {
	// BlockRef BlockRef
	// Nodes []SenderSignature
	// RandomSeedSignature primitives.Bls1Sig

	// internal
	// implements membuffers.Message
	_message membuffers.InternalMessage
}

func (x *BlockProof) String() string {
	if x == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{BlockRef:%s,Nodes:%s,RandomSeedSignature:%s,}", x.StringBlockRef(), x.StringNodes(), x.StringRandomSeedSignature())
}

var _BlockProof_Scheme = []membuffers.FieldType{membuffers.TypeMessage, membuffers.TypeMessageArray, membuffers.TypeBytes}
var _BlockProof_Unions = [][]membuffers.FieldType{}

func BlockProofReader(buf []byte) *BlockProof {
	x := &BlockProof{}
	x._message.Init(buf, membuffers.Offset(len(buf)), _BlockProof_Scheme, _BlockProof_Unions)
	return x
}

func (x *BlockProof) IsValid() bool {
	return x._message.IsValid()
}

func (x *BlockProof) Raw() []byte {
	return x._message.RawBuffer()
}

func (x *BlockProof) Equal(y *BlockProof) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return bytes.Equal(x.Raw(), y.Raw())
}

func (x *BlockProof) BlockRef() *BlockRef {
	b, s := x._message.GetMessage(0)
	return BlockRefReader(b[:s])
}

func (x *BlockProof) RawBlockRef() []byte {
	return x._message.RawBufferForField(0, 0)
}

func (x *BlockProof) RawBlockRefWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(0, 0)
}

func (x *BlockProof) StringBlockRef() string {
	return x.BlockRef().String()
}

func (x *BlockProof) NodesIterator() *BlockProofNodesIterator {
	return &BlockProofNodesIterator{iterator: x._message.GetMessageArrayIterator(1)}
}

type BlockProofNodesIterator struct {
	iterator *membuffers.Iterator
}

func (i *BlockProofNodesIterator) HasNext() bool {
	return i.iterator.HasNext()
}

func (i *BlockProofNodesIterator) NextNodes() *SenderSignature {
	b, s := i.iterator.NextMessage()
	return SenderSignatureReader(b[:s])
}

func (x *BlockProof) RawNodesArray() []byte {
	return x._message.RawBufferForField(1, 0)
}

func (x *BlockProof) RawNodesArrayWithHeader() []byte {
	return x._message.RawBufferWithHeaderForField(1, 0)
}

func (x *BlockProof) StringNodes() (res string) {
	res = "["
	for i := x.NodesIterator(); i.HasNext(); {
		res += i.NextNodes().String() + ","
	}
	res += "]"
	return
}

func (x *BlockProof) RandomSeedSignature() primitives.Bls1Sig {
	return primitives.Bls1Sig(x._message.GetBytes(2))
}

func (x *BlockProof) RawRandomSeedSignature() []byte {
	return x._message.RawBufferForField(2, 0)
}

func (x *BlockProof) MutateRandomSeedSignature(v primitives.Bls1Sig) error {
	return x._message.SetBytes(2, []byte(v))
}

func (x *BlockProof) StringRandomSeedSignature() string {
	return fmt.Sprintf("%s", x.RandomSeedSignature())
}

// builder

type BlockProofBuilder struct {
	BlockRef            *BlockRefBuilder
	Nodes               []*SenderSignatureBuilder
	RandomSeedSignature primitives.Bls1Sig

	// internal
	// implements membuffers.Builder
	_builder               membuffers.InternalBuilder
	_overrideWithRawBuffer []byte
}

func (w *BlockProofBuilder) arrayOfNodes() []membuffers.MessageWriter {
	res := make([]membuffers.MessageWriter, len(w.Nodes))
	for i, v := range w.Nodes {
		res[i] = v
	}
	return res
}

func (w *BlockProofBuilder) Write(buf []byte) (err error) {
	if w == nil {
		return
	}
	w._builder.NotifyBuildStart()
	defer w._builder.NotifyBuildEnd()
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	if w._overrideWithRawBuffer != nil {
		return w._builder.WriteOverrideWithRawBuffer(buf, w._overrideWithRawBuffer)
	}
	w._builder.Reset()
	err = w._builder.WriteMessage(buf, w.BlockRef)
	if err != nil {
		return
	}
	err = w._builder.WriteMessageArray(buf, w.arrayOfNodes())
	if err != nil {
		return
	}
	w._builder.WriteBytes(buf, []byte(w.RandomSeedSignature))
	return nil
}

func (w *BlockProofBuilder) HexDump(prefix string, offsetFromStart membuffers.Offset) (err error) {
	if w == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = &membuffers.ErrBufferOverrun{}
		}
	}()
	w._builder.Reset()
	err = w._builder.HexDumpMessage(prefix, offsetFromStart, "BlockProof.BlockRef", w.BlockRef)
	if err != nil {
		return
	}
	err = w._builder.HexDumpMessageArray(prefix, offsetFromStart, "BlockProof.Nodes", w.arrayOfNodes())
	if err != nil {
		return
	}
	w._builder.HexDumpBytes(prefix, offsetFromStart, "BlockProof.RandomSeedSignature", []byte(w.RandomSeedSignature))
	return nil
}

func (w *BlockProofBuilder) GetSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	return w._builder.GetSize()
}

func (w *BlockProofBuilder) CalcRequiredSize() membuffers.Offset {
	if w == nil {
		return 0
	}
	w.Write(nil)
	return w._builder.GetSize()
}

func (w *BlockProofBuilder) Build() *BlockProof {
	buf := make([]byte, w.CalcRequiredSize())
	if w.Write(buf) != nil {
		return nil
	}
	return BlockProofReader(buf)
}

func BlockProofBuilderFromRaw(raw []byte) *BlockProofBuilder {
	return &BlockProofBuilder{_overrideWithRawBuffer: raw}
}

/////////////////////////////////////////////////////////////////////////////
// enums

type MessageType uint16

const (
	LEAN_HELIX_RESERVED    MessageType = 0
	LEAN_HELIX_PREPREPARE  MessageType = 1
	LEAN_HELIX_PREPARE     MessageType = 2
	LEAN_HELIX_COMMIT      MessageType = 3
	LEAN_HELIX_NEW_VIEW    MessageType = 4
	LEAN_HELIX_VIEW_CHANGE MessageType = 5
)

func (n MessageType) String() string {
	switch n {
	case LEAN_HELIX_RESERVED:
		return "LEAN_HELIX_RESERVED"
	case LEAN_HELIX_PREPREPARE:
		return "LEAN_HELIX_PREPREPARE"
	case LEAN_HELIX_PREPARE:
		return "LEAN_HELIX_PREPARE"
	case LEAN_HELIX_COMMIT:
		return "LEAN_HELIX_COMMIT"
	case LEAN_HELIX_NEW_VIEW:
		return "LEAN_HELIX_NEW_VIEW"
	case LEAN_HELIX_VIEW_CHANGE:
		return "LEAN_HELIX_VIEW_CHANGE"
	}
	return "UNKNOWN"
}
