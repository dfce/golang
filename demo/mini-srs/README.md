# 直播 推流

## 
- OBS 推流（RTMP Publish）
- VLC/ffplay播放（RTMP Play）
- 多个播放器同时观看
- GOP Cache 秒开
- HTTP-FLV
- HLS
- 扩展WebRTC

## 结构图
SRS
↓
Source
Consumer
RTMP
AV
MediaMTX
↓
Path
Reader
RTMP
H264
```
mini-srs

cmd
└── server
    └── main.go


internal
├── bootstrap
│   └── app.go
├── config
│   └── config.go
├── server
│   └── session.go
├── protocol
│   └── rtmp
│       └── handshake.go
│       └── chunk_reader.go
│       └── chunk_writer.go
│       └── message.go
│   └── flv
│       └── video.go
│       └── audio.go
│   └── amf
│       └── decoder.go
│       └── encoder.go
├── stream
│   └── manager.go
│   └── stream.go
│   └── publisher.go
│   └── subscriber.go
│   └── go_cache.go
├── av
│   └── h264.go
│   └── aac.go

pkg
├── response
├── errno
├── jwt
└── constant
```

## 模拟Source推流 测试流程
```sh
# 1. 启动
go run cms/server/main.go

# 在当前命令执行目录生成测试视频
ffmpeg -f lavfi -i testsrc=duration=30:size=1280x720:rate=30 -f lavfi -i sine=frequency=1000:duration=30 -c:v libx264 -c:a aac test.mp4


# 2. 模拟Source推流
# 方案一：
# 使用 FFmpeg命令行（最推荐、最精准）
# 参数：
#   -re 按照视频真实帧率读取（不会瞬间推完）
#   -f flv 时RTMP必须的封装格式
ffmpeg -re -i "test.mp4" -vcodec libx264 -acodec aac -f flv rtmp://127.0.0.1:1935/live/test
# 用绝对的硬性命令，强制高版本的FFmpeg必须做直播推流，严格限制其每秒只能推1秒的数据
# 高版本FFmpeg引入了更加通用、绝对不会失效的实时流控制参数 -readrate 1(或标准位置的 -re)
# 参数：
#   -readrate 1 强制实时速率控制器，1=严格以1.0X的标准视频速率进行流式读取
#   -c:v h264_videotoolbox 调用 M1 Mac的硬解码，保证编码时钟的高精度对齐
ffmpeg -readrate 1 -i test.mp4 -c:v h264_videotoolbox -c:a aac -f flv rtmp://127.0.0.1:1935/live/test


# 方案二：使用OBS Studio(忽略)

# 3. 模拟 Consumer/Reader 拉流观看
# 新开终端
ffplay rtmp://127.0.0.1:1935/live/test

```