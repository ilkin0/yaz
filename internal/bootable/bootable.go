package bootable

func MakeBootable(device, label string) error {
	return createGptPartition(device, label)
}
