## TODO

### Done
- [x] "Flash another" flow after completion (return to home)
- [x] macOS support (diskutil for device scanning + unmount)
- [x] Bootable USB creation (UEFI)

### High priority
- [x] Win11 WIM splitting — split `sources/install.wim` (>4GB) into <4GB `.swm` chunks via `wimlib-imagex split`, single FAT32 partition. Windows installer reads split SWMs natively.
- [ ] Handle ctrl+c gracefully during write (cleanup, unmount)
- [ ] Error handling for edge cases (device removed mid-write, permission denied)

### Medium priority
- [ ] Hybrid ISO mode selection — let user choose file-copy (ISO Image) vs raw write (DD Image) for hybrid ISOs, like Rufus does. File-copy preserves drive usability, DD is more boot-compatible.
- [ ] Enter path manually (text input for image)
- [ ] Compression support (.gz, .xz — stream decompress while writing)

### Low priority
- [ ] Download image from URL
- [ ] Disk wiping (zero-fill / secure erase)
- [ ] Windows host support
