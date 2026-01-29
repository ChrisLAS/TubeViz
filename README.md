# TubeViz 🎬

> Transform your podcast audio into stunning YouTube visualizations with Synthwave-themed frequency bars and advanced algorithms.

Forked from [Jivefire](https://github.com/linuxmatters/jivefire) with enhanced visualization algorithms inspired by [vibeviz](https://github.com/noblepayne/vibeviz).

## The Groove

Your podcast audio deserves more than a static image on YouTube. TubeViz transforms WAV/MP3/FLAC into delightful 720p visuals—bars that breathe with your dialogue, rise with your laughter, and groove through every frequency.

<div align="center"><img alt="TubeViz Demo" src=".github/jivefire.gif" width="860" /></div>

### Enhanced Features

- 🌈 **Synthwave Theme** - Beautiful purple→pink→cyan gradient coloring
- 📊 **Logarithmic Frequency Binning** - Perceptually accurate 40Hz-18kHz distribution
- 🎯 **Spike Prevention** - Smooth visualization without audio artifacts  
- 🏃 **Velocity-Based Physics** - Natural peak motion with realistic decay
- 🖼️ **Thumbnail generator** YouTube-style PNG with your title, saved alongside the video
- 🎬 **1280×720 @ 30fps** H.264/AAC YouTube-ready MP4, no questions asked
  - 🎚️ **64 frequency bars** that actually look discrete (not that smeared spectrum nonsense)
  - 🪞 **Symmetric mirroring** above and below centre, doubles the visual impact
  - 🔬 **FFT-based analysis** 2048-point Hanning window, logarithmic frequency binning
  - ✨ **Smooth decay animation** - Velocity-based physics with natural motion
- 🚀 **Stupidly fast** streaming pipeline, parallel RGB→YUV conversion
  - ⚡ **GPU acceleration** auto-detected: NVENC, Vulkan, VA-API, QuickSync, VideoToolbox
- 📦 **Single binary** No Python. No FFmpeg install required. Just drop and render
  - 🐧 **Linux** (amd64 and aarch64)
  - 🍏 **macOS** (x86 and Apple Silicon)

## Usage

### Generate Video (Default Theme)
```bash
./tubeviz input.wav output.mp4 --theme default
```

### Generate Video (Synthwave Theme)
```bash
./tubeviz input.wav output.mp4 --theme synthwave
```

### With Episode Number and Title
```bash
./tubeviz --episode=42 --title="Linux Matters" input.wav output.mp4 --theme synthwave
```

### Example

<div align="center">
  <a href="https://www.youtube.com/watch?v=VPJEQhdaXrk" target="_blank">
    <img alt="Linux Matters: Episode 65 (macOS Made Me Snap)" src=".github/thumbnail.png" width="640">
  </a>
</div>

## Build

TubeViz uses [ffmpeg-statigo](https://github.com/linuxmatters/ffmpeg-statigo) for FFmpeg static bindings.

```bash
# Setup or update ffmpeg-statigo submodule and library
just setup

# Build and test
just build        # Build binary
just test         # Run tests
just test-encoder # Test encoder
```

## Why TubeViz?

FFmpeg's audio visualisation filters (`showfreqs`, `showspectrum`) render continuous frequency spectra, not discrete bars. No amount of FFmpeg filter chain kung-fu can achieve the discrete 64-bar aesthetic required for podcast branding. Solution: Do the FFT analysis and bar rendering in Go, pipe frames to FFmpeg for encoding.

**Why Go over Python?** The original `djfun/audio-visualizer-python` tool is a moribund Qt5 GUI with significant tech debt. For our podcast production needs we wanted multi-architecture tools that can integrate into automation pipelines.

The TubeViz architecture is available in the [ARCHITECTURE.md](docs/ARCHITECTURE.md) document.

## Attribution

TubeViz is based on [Jivefire](https://github.com/linuxmatters/jivefire) by Linux Matters, with visualization algorithm improvements inspired by [vibeviz](https://github.com/noblepayne/vibeviz). Licensed under GPL v3.
