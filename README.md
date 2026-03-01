# go-scrcpy

scrcpy v3.3.4 客户端协议的 Go 实现。提供视频/音频流数据包解析、帧拆解与缓存、控制消息构建、设备消息解析等全套功能，可作为依赖库集成到任何 Go 项目中。

无外部依赖，仅使用 Go 标准库。

## 安装

```bash
go get github.com/Jlan45/go-scrcpy
```

## 连接模型

scrcpy v2+ 使用**三条独立 TCP 连接**：

```
视频连接  →  握手读取  →  循环读取视频数据包
音频连接  →  握手读取  →  循环读取音频数据包
控制连接  →  握手读取  →  双向：发送控制消息 / 读取设备消息
```

每条连接建立后，服务端都会先发送 64 字节设备名。视频连接额外发送编解码器 ID 和初始分辨率，音频连接额外发送编解码器 ID。

使用 ADB 建立端口转发（通常是 `adb forward tcp:27183 localabstract:scrcpy`），然后用标准 `net.Dial` 连接，得到的 `net.Conn` 可以直接传入本库的所有函数。

## 快速上手

### 使用 Session 一步初始化（推荐）

`NewSession` 接受三条已建立的 TCP 连接，完成全部握手、启动所有后台 goroutine，并以结构体的形式返回。

```go
package main

import (
    "fmt"
    "log"
    "net"

    scrcpy "github.com/Jlan45/go-scrcpy"
)

func main() {
    // 假设已通过 adb forward 建立端口转发：
    //   adb forward tcp:27183 localabstract:scrcpy
    // scrcpy v2+ 按顺序建立三条连接（视频、音频、控制）
    dial := func() net.Conn {
        c, err := net.Dial("tcp", "127.0.0.1:27183")
        if err != nil {
            log.Fatal(err)
        }
        return c
    }

    videoConn := dial()
    audioConn := dial()
    ctrlConn  := dial()
    defer videoConn.Close()
    defer audioConn.Close()
    defer ctrlConn.Close()

    // 一次调用完成所有握手，启动所有后台 goroutine
    sess, err := scrcpy.NewSession(videoConn, audioConn, ctrlConn, scrcpy.SessionOptions{
        VideoBufSize: 30, // 视频帧缓冲数，默认 30
        AudioBufSize: 60, // 音频包缓冲数，默认 60
        CtrlBufSize:  16, // 设备消息缓冲数，默认 16
    })
    if err != nil {
        log.Fatal(err)
    }
    defer sess.Close()

    fmt.Println("设备:", sess.Device.DeviceName)
    fmt.Printf("视频: codec=%#x  %dx%d\n", sess.Video.Codec, sess.Video.Width, sess.Video.Height)
    fmt.Printf("音频: codec=%#x\n", sess.Audio.Codec)

    // ── 视频帧循环 ────────────────────────────────────────────────────────
    go func() {
        for frame := range sess.Frames.Frames() {
            if frame.Config != nil {
                // decoder.Init(frame.Config)
            }
            _ = frame // decoder.Decode(frame.PTS, frame.Data)
        }
    }()

    // ── 音频包循环 ────────────────────────────────────────────────────────
    go func() {
        for pkt := range sess.AudioPkts.Packets() {
            if pkt.IsConfig {
                // audioDecoder.Init(pkt.Data)
                continue
            }
            _ = pkt // audioDecoder.Decode(pkt.PTS, pkt.Data)
        }
    }()

    // ── 设备消息循环 ──────────────────────────────────────────────────────
    go func() {
        for msg := range sess.Control.DeviceMsgs() {
            if msg.Type == scrcpy.DeviceMsgClipboard {
                fmt.Println("剪贴板:", msg.ClipboardText)
            }
        }
    }()

    // ── 发送控制消息 ──────────────────────────────────────────────────────
    screen := scrcpy.Size{Width: uint16(sess.Video.Width), Height: uint16(sess.Video.Height)}
    sess.Control.InjectTouchEvent(
        scrcpy.MotionActionDown, scrcpy.PointerIDMouse,
        scrcpy.Point{X: 540, Y: 960}, screen, 1.0, 0, scrcpy.ButtonPrimary,
    )
    sess.Control.InjectTouchEvent(
        scrcpy.MotionActionUp, scrcpy.PointerIDMouse,
        scrcpy.Point{X: 540, Y: 960}, screen, 0, 0, 0,
    )
}
```

**禁用部分流时**（例如 `--no-audio`），对应连接传 `nil`，Session 中对应字段也为 `nil`：

```go
// 只有视频和控制，没有音频
sess, err := scrcpy.NewSession(videoConn, nil, ctrlConn, scrcpy.SessionOptions{})
// sess.AudioPkts == nil
// sess.Audio.Codec == 0 (CodecDisabled)
```

**`Session` 字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `Device` | `DeviceInfo` | 设备名（取自第一条成功握手的连接） |
| `Video` | `VideoInfo` | 视频编解码器 + 初始分辨率 |
| `Audio` | `AudioInfo` | 音频编解码器 |
| `Frames` | `*FrameDemuxer` | 视频帧 channel，`videoConn=nil` 时为 nil |
| `AudioPkts` | `*PacketBuffer` | 音频包 channel，`audioConn=nil` 时为 nil |
| `Control` | `*ControlConn` | 控制连接，`ctrlConn=nil` 时为 nil |

---

### 手动初始化（底层方式）

### 完整示例

```go
package main

import (
    "fmt"
    "net"

    scrcpy "github.com/Jlan45/go-scrcpy"
)

func main() {
    // 假设已通过 adb forward 建立端口转发
    videoConn, _ := net.Dial("tcp", "127.0.0.1:27183")
    audioConn, _ := net.Dial("tcp", "127.0.0.1:27183")
    ctrlConn, _  := net.Dial("tcp", "127.0.0.1:27183")
    defer videoConn.Close()
    defer audioConn.Close()
    defer ctrlConn.Close()

    // ── 视频连接握手 ──────────────────────────────────────────
    devInfo, _ := scrcpy.ReadDeviceInfo(videoConn)
    videoInfo, _ := scrcpy.ReadVideoInfo(videoConn)
    fmt.Printf("设备: %s  编解码器: %#x  分辨率: %dx%d\n",
        devInfo.DeviceName, videoInfo.Codec,
        videoInfo.Width, videoInfo.Height)

    // ── 音频连接握手 ──────────────────────────────────────────
    scrcpy.ReadDeviceInfo(audioConn)
    audioInfo, _ := scrcpy.ReadAudioInfo(audioConn)
    fmt.Printf("音频编解码器: %#x\n", audioInfo.Codec)

    // ── 控制连接握手 ──────────────────────────────────────────
    scrcpy.ReadDeviceInfo(ctrlConn)

    // ── 读取视频包（循环） ────────────────────────────────────
    go func() {
        for {
            pkt, err := scrcpy.ReadPacket(videoConn)
            if err != nil {
                return
            }
            if pkt.IsConfig {
                fmt.Println("收到视频配置包（SPS/PPS）")
                continue
            }
            fmt.Printf("视频帧 PTS=%dµs  keyframe=%v  size=%d\n",
                pkt.PTS, pkt.IsKeyframe, len(pkt.Data))
        }
    }()

    // ── 读取设备消息（循环） ──────────────────────────────────
    go func() {
        for {
            msg, err := scrcpy.ReadDeviceMsg(ctrlConn)
            if err != nil {
                return
            }
            switch msg.Type {
            case scrcpy.DeviceMsgClipboard:
                fmt.Println("剪贴板:", msg.ClipboardText)
            case scrcpy.DeviceMsgAckClipboard:
                fmt.Println("SET_CLIPBOARD 已确认, seq=", msg.AckSequence)
            }
        }
    }()

    // ── 发送控制消息 ──────────────────────────────────────────
    // 点击屏幕 (540, 960)
    screen := scrcpy.Size{Width: 1080, Height: 1920}
    pos    := scrcpy.Point{X: 540, Y: 960}

    ctrlConn.Write(scrcpy.BuildInjectTouchEvent(
        scrcpy.MotionActionDown, scrcpy.PointerIDMouse,
        pos, screen, 1.0, 0, scrcpy.ButtonPrimary,
    ))
    ctrlConn.Write(scrcpy.BuildInjectTouchEvent(
        scrcpy.MotionActionUp, scrcpy.PointerIDMouse,
        pos, screen, 0, 0, 0,
    ))
}
```

---

## API 参考

### 握手读取

#### `ReadDeviceInfo(r io.Reader) (*DeviceInfo, error)`

读取每条连接开头的 64 字节设备名。必须在其他任何读取之前调用。

```go
info, err := scrcpy.ReadDeviceInfo(conn)
fmt.Println(info.DeviceName) // e.g. "Pixel 8 Pro"
```

#### `ReadVideoInfo(r io.Reader) (*VideoInfo, error)`

在视频连接上，紧接 `ReadDeviceInfo` 之后调用。返回编解码器 ID 和初始分辨率。

```go
v, err := scrcpy.ReadVideoInfo(videoConn)
// v.Codec  — CodecH264 / CodecH265 / CodecAV1
// v.Width, v.Height — 初始编码分辨率（旋转后通过 config 包更新）
```

如果 `v.Codec == scrcpy.CodecDisabled`，服务端以 `--no-video` 启动。

#### `ReadAudioInfo(r io.Reader) (*AudioInfo, error)`

在音频连接上，紧接 `ReadDeviceInfo` 之后调用。

```go
a, err := scrcpy.ReadAudioInfo(audioConn)
// a.Codec — CodecOpus / CodecAAC / CodecFLAC / CodecRAW / CodecDisabled
```

---

### 视频帧拆解与缓存

#### `NewFrameDemuxer(r io.Reader, bufSize int) *FrameDemuxer`

在后台 goroutine 中持续读取视频数据包，自动将 codec 配置包（SPS/PPS 等）与紧随其后的数据帧合并，通过带缓冲 channel 交付 `VideoFrame`。

```go
// 启动帧拆解器，缓冲 30 帧
d := scrcpy.NewFrameDemuxer(videoConn, 30)
defer d.Close()

for frame := range d.Frames() {
    if frame.Config != nil {
        // codec 配置发生变化（流开始或屏幕旋转）
        // 用 frame.Config 重新初始化解码器
        decoder.Init(frame.Config)
    }
    // frame.Data 是一个完整的编码帧（H.264 NAL / H.265 NAL / AV1 OBU）
    decoder.Decode(frame.PTS, frame.Data)
}

// channel 关闭后检查错误
if err := d.Err(); err != nil {
    log.Println("demuxer stopped:", err)
}
```

**`VideoFrame` 字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `PTS` | `int64` | 显示时间戳（微秒） |
| `IsKeyframe` | `bool` | 是否为关键帧（I 帧） |
| `Config` | `[]byte` | codec 配置变化时非 nil（H.264 SPS+PPS / H.265 VPS+SPS+PPS / AV1 序列头） |
| `Data` | `[]byte` | 编码帧负载 |

**FrameDemuxer 方法：**

| 方法 | 说明 |
|------|------|
| `Frames() <-chan *VideoFrame` | 返回帧 channel，channel 关闭表示流结束或出错 |
| `Err() error` | channel 关闭后获取停止原因，正常 Close 返回 nil |
| `Len() int` | 当前缓冲中的帧数 |
| `Cap() int` | 缓冲区容量（构造时指定） |
| `Close()` | 通知 goroutine 退出；**不会**关闭底层连接 |

> **注意**：`Close()` 仅发送停止信号，不会中断正在进行的 `ReadPacket` 网络阻塞调用。
> 若要立即停止，需同时关闭底层的 `net.Conn`：`videoConn.Close()`。

**缓冲策略**：`bufSize` 为帧队列容量。消费方跟不上时，生产方会阻塞等待（背压），不会丢帧。`bufSize=0` 为无缓冲模式，生产和消费完全同步。

**屏幕旋转处理示例：**

```go
var decoder *MyDecoder

for frame := range d.Frames() {
    if frame.Config != nil {
        // 旧解码器作废，用新的 SPS/PPS 重建
        decoder.Close()
        decoder = NewDecoder(frame.Config)
    }
    decoder.Decode(frame.PTS, frame.Data)
}
```

---

### 数据包读取（底层）

#### `ReadPacket(r io.Reader) (*Packet, error)`

从视频或音频连接中读取一个数据包。阻塞直到收到完整包。

```go
type Packet struct {
    PTS        int64  // 显示时间戳（微秒），IsConfig=true 时无意义
    IsConfig   bool   // true = 编解码器配置包（H.264 SPS/PPS 等），须优先送给解码器
    IsKeyframe bool   // true = 关键帧（I 帧）
    Data       []byte // 原始编码负载
}
```

典型处理逻辑：

```go
for {
    pkt, err := scrcpy.ReadPacket(videoConn)
    if err != nil {
        break
    }
    if pkt.IsConfig {
        decoder.SendConfig(pkt.Data)  // 初始化解码器
    } else {
        decoder.SendFrame(pkt.PTS, pkt.Data)
    }
}
```

---

### 控制连接封装

#### `NewControlConn(rw io.ReadWriter, bufSize int) *ControlConn`

将控制 socket 封装为 `ControlConn`，解决两个核心问题：

1. **并发写安全**：内置 `sync.Mutex`，任意 goroutine 可同时调用 `Send` 或任意方法，不会产生帧交叉。
2. **双向分离**：后台 goroutine 负责读取设备消息并推入 channel，与发送方向完全解耦。

```go
// 握手完成后构建 ControlConn，bufSize 为设备消息缓冲数量
ctrl := scrcpy.NewControlConn(ctrlConn, 16)
defer ctrl.Close()

// 在独立 goroutine 接收设备消息
go func() {
    for msg := range ctrl.DeviceMsgs() {
        switch msg.Type {
        case scrcpy.DeviceMsgClipboard:
            fmt.Println("设备剪贴板:", msg.ClipboardText)
        case scrcpy.DeviceMsgAckClipboard:
            fmt.Println("SET_CLIPBOARD 已确认, seq=", msg.AckSequence)
        case scrcpy.DeviceMsgUHIDOutput:
            fmt.Printf("UHID #%d 输出: %x\n", msg.UHID_ID, msg.UHID_Data)
        }
    }
    if err := ctrl.Err(); err != nil {
        log.Println("control conn error:", err)
    }
}()

// 任意 goroutine 中安全地发送
screen := scrcpy.Size{Width: 1080, Height: 1920}
ctrl.InjectTouchEvent(
    scrcpy.MotionActionDown, scrcpy.PointerIDMouse,
    scrcpy.Point{X: 540, Y: 960},
    screen, 1.0, 0, scrcpy.ButtonPrimary,
)
ctrl.InjectTouchEvent(
    scrcpy.MotionActionUp, scrcpy.PointerIDMouse,
    scrcpy.Point{X: 540, Y: 960},
    screen, 0, 0, 0,
)
ctrl.InjectText("Hello")
ctrl.SetClipboard(1, true, "粘贴内容")
ctrl.SetDisplayPower(false)
```

**`ControlConn` 方法一览：**

| 分类 | 方法 |
|------|------|
| 底层 | `Send([]byte) error`、`DeviceMsgs() <-chan *DeviceMsg`、`Err() error`、`Close()` |
| 键盘 | `InjectKeycode`、`InjectText` |
| 触摸/鼠标 | `InjectTouchEvent`、`InjectScrollEvent` |
| 系统 | `BackOrScreenOn`、`ExpandNotificationPanel`、`ExpandSettingsPanel`、`CollapsePanels`、`RotateDevice`、`SetDisplayPower`、`ResetVideo`、`OpenHardKeyboardSettings` |
| 剪贴板 | `GetClipboard`、`SetClipboard` |
| 应用 | `StartApp` |
| HID | `UHIDCreate`、`UHIDInput`、`UHIDDestroy` |

> `Close()` 只发送停止信号，不关闭底层连接。要立即中断阻塞中的读取，需同时调用 `ctrlConn.Close()`。

---

### 控制消息（底层函数）

所有 `Build*` 函数返回 `[]byte`，可直接写入连接，或通过 `ControlConn.Send()` 发送。

#### 键盘输入

```go
// 按下 HOME 键
ctrlConn.Write(scrcpy.BuildInjectKeycode(
    scrcpy.KeyEventActionDown, scrcpy.KeycodeHome, 0, 0,
))
ctrlConn.Write(scrcpy.BuildInjectKeycode(
    scrcpy.KeyEventActionUp, scrcpy.KeycodeHome, 0, 0,
))

// 带修饰键：Ctrl+C
ctrlConn.Write(scrcpy.BuildInjectKeycode(
    scrcpy.KeyEventActionDown, scrcpy.KeycodeC, 0,
    scrcpy.MetaCtrlLeftOn|scrcpy.MetaCtrlOn,
))
```

常用 KeyCode：`KeycodeHome`、`KeycodeBack`、`KeycodeAppSwitch`、`KeycodePower`、`KeycodeVolumeUp`、`KeycodeVolumeDown`、`KeycodeEnter`、`KeycodeDel`（退格）、`KeycodeForwardDel`（Delete）、`KeycodeEscape`、`KeycodeA`…`KeycodeZ`、`Keycode0`…`Keycode9`、`KeycodeF1`…`KeycodeF12`。

#### 文本输入

```go
// 直接输入文本（不走键盘映射，适合中文等非 ASCII 字符）
ctrlConn.Write(scrcpy.BuildInjectText("Hello, 世界"))
```

#### 触摸 / 鼠标

```go
screen := scrcpy.Size{Width: 1080, Height: 1920}

// 单次点击
down := scrcpy.BuildInjectTouchEvent(
    scrcpy.MotionActionDown,
    scrcpy.PointerIDMouse,       // 鼠标光标用 PointerIDMouse，手指用 0、1、2…
    scrcpy.Point{X: 300, Y: 500},
    screen,
    1.0,                          // 压力 [0.0, 1.0]
    scrcpy.ButtonPrimary,         // actionButton（触发本次动作的按键）
    scrcpy.ButtonPrimary,         // buttons（当前所有按下的按键）
)
up := scrcpy.BuildInjectTouchEvent(
    scrcpy.MotionActionUp,
    scrcpy.PointerIDMouse,
    scrcpy.Point{X: 300, Y: 500},
    screen, 0, 0, 0,
)
ctrlConn.Write(down)
ctrlConn.Write(up)

// 右键单击
ctrlConn.Write(scrcpy.BuildInjectTouchEvent(
    scrcpy.MotionActionDown, scrcpy.PointerIDMouse,
    scrcpy.Point{X: 300, Y: 500}, screen,
    1.0, scrcpy.ButtonSecondary, scrcpy.ButtonSecondary,
))
```

多点触控示例（双指缩放）：

```go
// 手指 0 按下
ctrlConn.Write(scrcpy.BuildInjectTouchEvent(
    scrcpy.MotionActionDown, 0,
    scrcpy.Point{X: 400, Y: 800}, screen, 1.0, 0, 0,
))
// 手指 1 按下（PointerDown）
ctrlConn.Write(scrcpy.BuildInjectTouchEvent(
    scrcpy.MotionActionPointerDown, 1,
    scrcpy.Point{X: 700, Y: 800}, screen, 1.0, 0, 0,
))
// 两指移动（发两条 MOVE）……
// 手指 1 抬起
ctrlConn.Write(scrcpy.BuildInjectTouchEvent(
    scrcpy.MotionActionPointerUp, 1,
    scrcpy.Point{X: 800, Y: 800}, screen, 0, 0, 0,
))
// 手指 0 抬起
ctrlConn.Write(scrcpy.BuildInjectTouchEvent(
    scrcpy.MotionActionUp, 0,
    scrcpy.Point{X: 300, Y: 800}, screen, 0, 0, 0,
))
```

#### 滚动

```go
// 向下滚动（vScroll 为负表示向下）
ctrlConn.Write(scrcpy.BuildInjectScrollEvent(
    scrcpy.Point{X: 540, Y: 960},
    screen,
    0,    // hScroll [-1.0, 1.0]
    -1.0, // vScroll [-1.0, 1.0]
    0,
))
```

#### 系统操作

```go
ctrlConn.Write(scrcpy.BuildBackOrScreenOn(scrcpy.KeyEventActionDown))
ctrlConn.Write(scrcpy.BuildBackOrScreenOn(scrcpy.KeyEventActionUp))

ctrlConn.Write(scrcpy.BuildExpandNotificationPanel())
ctrlConn.Write(scrcpy.BuildExpandSettingsPanel())
ctrlConn.Write(scrcpy.BuildCollapsePanels())
ctrlConn.Write(scrcpy.BuildRotateDevice())
ctrlConn.Write(scrcpy.BuildSetDisplayPower(false)) // 关屏
ctrlConn.Write(scrcpy.BuildSetDisplayPower(true))  // 开屏
ctrlConn.Write(scrcpy.BuildOpenHardKeyboardSettings())
ctrlConn.Write(scrcpy.BuildResetVideo())
```

#### 剪贴板

```go
// 读取设备剪贴板（服务端推送 DeviceMsgClipboard）
ctrlConn.Write(scrcpy.BuildGetClipboard(scrcpy.CopyKeyCopy))

// 写入设备剪贴板，序号用于匹配 ACK_CLIPBOARD 响应
var seq uint64 = 1
ctrlConn.Write(scrcpy.BuildSetClipboard(seq, true, "粘贴这段文字"))
```

#### 启动应用

```go
ctrlConn.Write(scrcpy.BuildStartApp("com.android.settings"))
```

#### 虚拟 HID 设备（UHID）

```go
const keyboardID uint16 = 1

// 注册虚拟键盘
ctrlConn.Write(scrcpy.BuildUHIDCreate(
    keyboardID,
    0x1234, 0x5678,       // vendorID, productID
    "Virtual Keyboard",
    reportDescriptor,     // []byte，标准 HID report descriptor
))

// 发送 HID 输入报告
ctrlConn.Write(scrcpy.BuildUHIDInput(keyboardID, reportData))

// 注销
ctrlConn.Write(scrcpy.BuildUHIDDestroy(keyboardID))
```

---

### 设备消息（服务端 → 客户端）

#### `ReadDeviceMsg(r io.Reader) (*DeviceMsg, error)`

从控制连接读取一条设备消息，阻塞直到收到完整消息。

```go
msg, err := scrcpy.ReadDeviceMsg(ctrlConn)
if err != nil {
    // 连接断开或未知消息类型
}

switch msg.Type {
case scrcpy.DeviceMsgClipboard:
    // 设备剪贴板内容变化（响应 GET_CLIPBOARD 或设备主动推送）
    fmt.Println(msg.ClipboardText)

case scrcpy.DeviceMsgAckClipboard:
    // SET_CLIPBOARD 被确认，msg.AckSequence 与发送时的 sequence 对应
    fmt.Println("已确认 seq:", msg.AckSequence)

case scrcpy.DeviceMsgUHIDOutput:
    // 虚拟 HID 设备的输出报告（如键盘 LED 状态）
    fmt.Printf("UHID #%d 输出: %x\n", msg.UHID_ID, msg.UHID_Data)
}
```

---

## 编解码器常量

| 常量 | 类型 | 说明 |
|------|------|------|
| `CodecH264` | 视频 | H.264 / AVC |
| `CodecH265` | 视频 | H.265 / HEVC |
| `CodecAV1`  | 视频 | AV1 |
| `CodecOpus` | 音频 | Opus |
| `CodecAAC`  | 音频 | AAC |
| `CodecFLAC` | 音频 | FLAC（无损） |
| `CodecRAW`  | 音频 | 原始 PCM（双声道 48 kHz） |
| `CodecDisabled` | — | 该流已禁用（`--no-video` / `--no-audio`） |
| `CodecError`    | — | 服务端编解码器协商失败 |

---

## 类型速查

```go
// 坐标和尺寸
type Point struct{ X, Y int32 }
type Size  struct{ Width, Height uint16 }

// 握手信息
type DeviceInfo struct{ DeviceName string }
type VideoInfo  struct{ Codec CodecID; Width, Height uint32 }
type AudioInfo  struct{ Codec CodecID }

// 数据包
type Packet struct {
    PTS        int64
    IsConfig   bool
    IsKeyframe bool
    Data       []byte
}

// 设备消息
type DeviceMsg struct {
    Type          DeviceMsgType
    ClipboardText string   // DeviceMsgClipboard
    AckSequence   uint64   // DeviceMsgAckClipboard
    UHID_ID       uint16   // DeviceMsgUHIDOutput
    UHID_Data     []byte   // DeviceMsgUHIDOutput
}
```

---

## 注意事项

- **协议版本**：本库实现 scrcpy **v3.3.4** 协议。scrcpy 协议属于内部协议，不保证跨版本兼容，客户端库版本须与服务端版本匹配。
- **字节序**：协议所有多字节整数均为**大端序**（网络字节序），本库已处理。
- **并发**：视频、音频、控制三条连接均可在独立 goroutine 中并发读写，互不干扰。
- **分辨率变化**：设备旋转时，`ReadPacket` 会返回一个新的 `IsConfig=true` 的包，其中携带新的 SPS（H.264）或等效配置，需重新初始化解码器。
- **控制连接的双向性**：控制连接既要发送控制消息（写），又要接收设备消息（读），建议分别用两个 goroutine 处理读和写。
