# yaz

A terminal UI tool for flashing OS images to USB drives. Built with Go and [Charm](https://charm.sh) libraries.

Supports Linux and macOS.

## Demo

![yaz demo](https://i.imgur.com/lBiQ4fe.gif)

## Features

- Flash ISO and raw disk images (.img) to USB drives
- Auto-detect hybrid vs non-hybrid ISOs
- UEFI bootable USB creation (GPT + FAT32)
- Windows 11 support — automatic WIM splitting for >4GB files
- Real-time progress bar with speed and ETA
- Device scanning and selection

## Requirements

- **Linux**: `parted`, `mkfs.vfat` (usually pre-installed)
- **macOS**: `diskutil` (built-in)
- **Windows 11 ISOs**: `wimlib-imagex` for WIM splitting
  - Debian/Ubuntu: `sudo apt install wimtools`
  - Fedora: `sudo dnf install wimlib-utils`
  - macOS: `brew install wimlib`

## Installation

### Homebrew (macOS/Linux)

```bash
brew install ilkin0/tap/yaz
```

### Go

```bash
go install github.com/ilkin0/yaz/cmd/yaz@latest
```

### From source

Requires [Go](https://go.dev/dl/) 1.25+ and `make`.

```bash
git clone https://github.com/ilkin0/yaz.git
cd yaz
make build
sudo ./bin/yaz
```

## Usage

```
sudo yaz
```

Root/sudo is required for device access.

## Notes

### Win11 USB boot

**The problem:** Win11 ISOs contain `sources/install.wim` (often >4GB). UEFI firmware only boots from FAT32, but FAT32 has a 4GB file size limit. So the biggest file in the ISO can't fit on the only bootable filesystem.

**Attempt 1 — Dual partition (FAT32 + NTFS):** Created a small FAT32 partition for boot files and a large NTFS partition for data. Failed twice:

1. First try: boot files on FAT32, everything else on NTFS → WIN RECOVERY error (BCD couldn't find `boot.wim`)
2. Second try: everything on FAT32, only >4GB files on NTFS → "Select driver to install" error (Windows PE couldn't find `install.wim` across partitions)

**Root cause:** Windows boot chain expects all files on a single volume. No major tool uses dual-partition — it fundamentally doesn't work without a custom UEFI driver.

**Why not NTFS?** Rufus uses single NTFS + a custom UEFI:NTFS driver (GPLv2 blob) that chainloads from NTFS.

**Why not raw-write (dd)?** Win11 ISOs are non-hybrid (no embedded partition table). Tools like balenaEtcher and Caligula only do raw writes, so they can't handle Windows ISOs at all.

**WIM splitting:** Single FAT32 partition. Copy all files normally, then split oversized `.wim` files into <4GB `.swm` chunks via `wimlib-imagex split`. Windows installer reads split SWMs natively. This is the same approach used by WoeUSB and windows2usb. Works with Secure Boot, no binary blobs, maximum compatibility.

### ISO types

- **Hybrid** (Linux distros): has MBR/GPT, can be dd'd or file-copied
- **Non-hybrid** (Win11): no partition table, must be file-copied with tool-created partitions
- **Raw .img** (RaspiOS): full partition table, goes through dd path

### Platform notes

- `wimlib-imagex`: `wimtools` (Debian/Ubuntu), `wimlib-utils` (Fedora), `wimlib` (macOS/Homebrew)
- ntfs-3g FUSE is ~2 MB/s write speed — irrelevant now since we use FAT32 only
