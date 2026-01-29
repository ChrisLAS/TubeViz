# TubeViz 🎬

> Transform your podcast audio into stunning YouTube visualizations with Synthwave-themed frequency bars and advanced algorithms.

Forked from [Jivefire](https://github.com/linuxmatters/jivefire) with enhanced visualization algorithms inspired by [vibeviz](https://github.com/noblepayne/vibeviz).

## The Groove

Your podcast audio deserves more than a static image on YouTube. TubeViz transforms WAV/MP3/FLAC into delightful 720p visuals—bars that breathe with your dialogue, rise with your laughter, and groove through every frequency.

<div align="center"><img alt="TubeViz Demo" src=".github/jivefire.gif" width="860" /></div>

### Enhanced Features

- 🎬 **Two-pass streaming pipeline** for accurate scaling with low memory use
- 🌈 **Synthwave theme** with purple→pink→cyan gradient bars
- 🎚️ **64 discrete bars** with logarithmic 40Hz–18kHz binning
- 🏃 **Velocity-based decay** that prevents spikes and keeps motion natural
- 🖼️ **Thumbnail generator** for YouTube-style episode art
- ⚡ **Parallel RGB→YUV conversion** with auto-detected GPU acceleration
- 📦 **Single static binary** with bundled FFmpeg (Linux amd64/aarch64, macOS Intel/Apple Silicon)

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
  <img alt="Linux Matters: Episode 65 (macOS Made Me Snap)" src=".github/thumbnail.png" width="640">
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
