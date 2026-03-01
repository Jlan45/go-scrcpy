package scrcpy

// ---------------------------------------------------------------------------
// Key event action
// ---------------------------------------------------------------------------

// KeyEventAction is the action field of an Android key event.
type KeyEventAction uint8

const (
	KeyEventActionDown     KeyEventAction = 0 // Key pressed
	KeyEventActionUp       KeyEventAction = 1 // Key released
	KeyEventActionMultiple KeyEventAction = 2 // Multiple key events (deprecated)
)

// ---------------------------------------------------------------------------
// Key codes  (android.view.KeyEvent KEYCODE_*)
// ---------------------------------------------------------------------------

// KeyCode is an Android key code.
type KeyCode uint32

const (
	KeycodeUnknown          KeyCode = 0
	KeycodeHome             KeyCode = 3
	KeycodeBack             KeyCode = 4
	KeycodeCall             KeyCode = 5
	KeycodeEndcall          KeyCode = 6
	Keycode0                KeyCode = 7
	Keycode1                KeyCode = 8
	Keycode2                KeyCode = 9
	Keycode3                KeyCode = 10
	Keycode4                KeyCode = 11
	Keycode5                KeyCode = 12
	Keycode6                KeyCode = 13
	Keycode7                KeyCode = 14
	Keycode8                KeyCode = 15
	Keycode9                KeyCode = 16
	KeycodeStar             KeyCode = 17
	KeycodePound            KeyCode = 18
	KeycodeDpadUp           KeyCode = 19
	KeycodeDpadDown         KeyCode = 20
	KeycodeDpadLeft         KeyCode = 21
	KeycodeDpadRight        KeyCode = 22
	KeycodeDpadCenter       KeyCode = 23
	KeycodeVolumeUp         KeyCode = 24
	KeycodeVolumeDown       KeyCode = 25
	KeycodePower            KeyCode = 26
	KeycodeCamera           KeyCode = 27
	KeycodeClear            KeyCode = 28
	KeycodeA                KeyCode = 29
	KeycodeB                KeyCode = 30
	KeycodeC                KeyCode = 31
	KeycodeD                KeyCode = 32
	KeycodeE                KeyCode = 33
	KeycodeF                KeyCode = 34
	KeycodeG                KeyCode = 35
	KeycodeH                KeyCode = 36
	KeycodeI                KeyCode = 37
	KeycodeJ                KeyCode = 38
	KeycodeK                KeyCode = 39
	KeycodeL                KeyCode = 40
	KeycodeM                KeyCode = 41
	KeycodeN                KeyCode = 42
	KeycodeO                KeyCode = 43
	KeycodeP                KeyCode = 44
	KeycodeQ                KeyCode = 45
	KeycodeR                KeyCode = 46
	KeycodeS                KeyCode = 47
	KeycodeT                KeyCode = 48
	KeycodeU                KeyCode = 49
	KeycodeV                KeyCode = 50
	KeycodeW                KeyCode = 51
	KeycodeX                KeyCode = 52
	KeycodeY                KeyCode = 53
	KeycodeZ                KeyCode = 54
	KeycodeComma            KeyCode = 55
	KeycodePeriod           KeyCode = 56
	KeycodeAltLeft          KeyCode = 57
	KeycodeAltRight         KeyCode = 58
	KeycodeShiftLeft        KeyCode = 59
	KeycodeShiftRight       KeyCode = 60
	KeycodeTab              KeyCode = 61
	KeycodeSpace            KeyCode = 62
	KeycodeSym              KeyCode = 63
	KeycodeExplorer         KeyCode = 64
	KeycodeEnvelope         KeyCode = 65
	KeycodeEnter            KeyCode = 66
	KeycodeDel              KeyCode = 67 // Backspace
	KeycodeGrave            KeyCode = 68
	KeycodeMinus            KeyCode = 69
	KeycodeEquals           KeyCode = 70
	KeycodeLeftBracket      KeyCode = 71
	KeycodeRightBracket     KeyCode = 72
	KeycodeBackslash        KeyCode = 73
	KeycodeSemicolon        KeyCode = 74
	KeycodeApostrophe       KeyCode = 75
	KeycodeSlash            KeyCode = 76
	KeycodeAt               KeyCode = 77
	KeycodeNum              KeyCode = 78
	KeycodeHeadsethook      KeyCode = 79
	KeycodeFocus            KeyCode = 80
	KeycodePlus             KeyCode = 81
	KeycodeMenu             KeyCode = 82
	KeycodeNotification     KeyCode = 83
	KeycodeSearch           KeyCode = 84
	KeycodeMediaPlayPause   KeyCode = 85
	KeycodeMediaStop        KeyCode = 86
	KeycodeMediaNext        KeyCode = 87
	KeycodeMediaPrevious    KeyCode = 88
	KeycodeMediaRewind      KeyCode = 89
	KeycodeMediaFastForward KeyCode = 90
	KeycodeMute             KeyCode = 91
	KeycodePageUp           KeyCode = 92
	KeycodePageDown         KeyCode = 93
	KeycodeEscape           KeyCode = 111
	KeycodeForwardDel       KeyCode = 112 // Delete
	KeycodeCtrlLeft         KeyCode = 113
	KeycodeCtrlRight        KeyCode = 114
	KeycodeCapsLock         KeyCode = 115
	KeycodeScrollLock       KeyCode = 116
	KeycodeMetaLeft         KeyCode = 117
	KeycodeMetaRight        KeyCode = 118
	KeycodeFunction         KeyCode = 119
	KeycodeSysrq            KeyCode = 120
	KeycodeBreak            KeyCode = 121
	KeycodeMoveHome         KeyCode = 122
	KeycodeMoveEnd          KeyCode = 123
	KeycodeInsert           KeyCode = 124
	KeycodeForward          KeyCode = 125
	KeycodeMediaPlay        KeyCode = 126
	KeycodeMediaPause       KeyCode = 127
	KeycodeMediaClose       KeyCode = 128
	KeycodeMediaEject       KeyCode = 129
	KeycodeMediaRecord      KeyCode = 130
	KeycodeF1               KeyCode = 131
	KeycodeF2               KeyCode = 132
	KeycodeF3               KeyCode = 133
	KeycodeF4               KeyCode = 134
	KeycodeF5               KeyCode = 135
	KeycodeF6               KeyCode = 136
	KeycodeF7               KeyCode = 137
	KeycodeF8               KeyCode = 138
	KeycodeF9               KeyCode = 139
	KeycodeF10              KeyCode = 140
	KeycodeF11              KeyCode = 141
	KeycodeF12              KeyCode = 142
	KeycodeNumLock          KeyCode = 143
	KeycodeNumpad0          KeyCode = 144
	KeycodeNumpad1          KeyCode = 145
	KeycodeNumpad2          KeyCode = 146
	KeycodeNumpad3          KeyCode = 147
	KeycodeNumpad4          KeyCode = 148
	KeycodeNumpad5          KeyCode = 149
	KeycodeNumpad6          KeyCode = 150
	KeycodeNumpad7          KeyCode = 151
	KeycodeNumpad8          KeyCode = 152
	KeycodeNumpad9          KeyCode = 153
	KeycodeNumpadDivide     KeyCode = 154
	KeycodeNumpadMultiply   KeyCode = 155
	KeycodeNumpadSubtract   KeyCode = 156
	KeycodeNumpadAdd        KeyCode = 157
	KeycodeNumpadDot        KeyCode = 158
	KeycodeNumpadComma      KeyCode = 159
	KeycodeNumpadEnter      KeyCode = 160
	KeycodeNumpadEquals     KeyCode = 161
	KeycodeNumpadLeftParen  KeyCode = 162
	KeycodeNumpadRightParen KeyCode = 163
	KeycodeVolumeMute       KeyCode = 164
	KeycodeInfo             KeyCode = 165
	KeycodeAppSwitch        KeyCode = 187
	KeycodeBrightnessDown   KeyCode = 220
	KeycodeBrightnessUp     KeyCode = 221
	KeycodeSleep            KeyCode = 223
	KeycodeWakeup           KeyCode = 224
	KeycodeAssist           KeyCode = 231
	KeycodeScreenshot       KeyCode = 120 // Some devices use this
)

// ---------------------------------------------------------------------------
// Meta state flags  (android.view.KeyEvent META_*)
// ---------------------------------------------------------------------------

// MetaState holds Android modifier key state bit flags.
type MetaState uint32

const (
	MetaShiftOn      MetaState = 0x00000001
	MetaAltOn        MetaState = 0x00000002
	MetaSymOn        MetaState = 0x00000004
	MetaFunctionOn   MetaState = 0x00000008
	MetaAltLeftOn    MetaState = 0x00000010
	MetaAltRightOn   MetaState = 0x00000020
	MetaShiftLeftOn  MetaState = 0x00000040
	MetaShiftRightOn MetaState = 0x00000080
	MetaCtrlOn       MetaState = 0x00001000
	MetaCtrlLeftOn   MetaState = 0x00002000
	MetaCtrlRightOn  MetaState = 0x00004000
	MetaMetaOn       MetaState = 0x00010000
	MetaMetaLeftOn   MetaState = 0x00020000
	MetaMetaRightOn  MetaState = 0x00040000
	MetaCapsLockOn   MetaState = 0x00100000
	MetaNumLockOn    MetaState = 0x00200000
	MetaScrollLockOn MetaState = 0x00400000
)

// ---------------------------------------------------------------------------
// Motion event action  (android.view.MotionEvent ACTION_*)
// ---------------------------------------------------------------------------

// MotionEventAction is the action field of an Android motion event.
type MotionEventAction uint8

const (
	MotionActionDown        MotionEventAction = 0
	MotionActionUp          MotionEventAction = 1
	MotionActionMove        MotionEventAction = 2
	MotionActionCancel      MotionEventAction = 3
	MotionActionOutside     MotionEventAction = 4
	MotionActionPointerDown MotionEventAction = 5
	MotionActionPointerUp   MotionEventAction = 6
	MotionActionHoverMove   MotionEventAction = 7
	MotionActionScroll      MotionEventAction = 8
	MotionActionHoverEnter  MotionEventAction = 9
	MotionActionHoverExit   MotionEventAction = 10
)

// ---------------------------------------------------------------------------
// Motion event buttons  (android.view.MotionEvent BUTTON_*)
// ---------------------------------------------------------------------------

// MotionEventButtons is a bitmask of pressed mouse buttons.
type MotionEventButtons uint32

const (
	ButtonPrimary   MotionEventButtons = 0x00000001 // Left mouse button
	ButtonSecondary MotionEventButtons = 0x00000002 // Right mouse button
	ButtonTertiary  MotionEventButtons = 0x00000004 // Middle mouse button
	ButtonBack      MotionEventButtons = 0x00000008 // Back button
	ButtonForward   MotionEventButtons = 0x00000010 // Forward button
	ButtonStylus    MotionEventButtons = 0x00000040 // Primary stylus button
	ButtonStylus2   MotionEventButtons = 0x00000080 // Secondary stylus button
)

// ---------------------------------------------------------------------------
// Special pointer IDs
// ---------------------------------------------------------------------------

// Well-known pointer IDs for use with BuildInjectTouchEvent.
const (
	// PointerIDMouse is the canonical pointer ID for a virtual mouse cursor.
	PointerIDMouse int64 = -1
	// PointerIDVirtualFinger is used for the second finger in a virtual pinch
	// gesture.
	PointerIDVirtualFinger int64 = -2
)

// ---------------------------------------------------------------------------
// GET_CLIPBOARD copy key
// ---------------------------------------------------------------------------

// GetClipboardCopyKey selects how the clipboard content is obtained.
type GetClipboardCopyKey uint8

const (
	CopyKeyNone GetClipboardCopyKey = 0 // Do not copy, just read
	CopyKeyCopy GetClipboardCopyKey = 1 // Trigger a copy operation first
	CopyKeyCut  GetClipboardCopyKey = 2 // Trigger a cut operation first
)

// ---------------------------------------------------------------------------
// Display power mode
// ---------------------------------------------------------------------------

// DisplayPowerMode controls the screen state for SET_DISPLAY_POWER.
type DisplayPowerMode uint8

const (
	DisplayPowerModeOff DisplayPowerMode = 0
	DisplayPowerModeOn  DisplayPowerMode = 2
)
