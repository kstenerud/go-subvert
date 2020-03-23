package subvert

func osGetMemoryProtection(address uintptr) memProtect {
	// TODO: Parse /proc/self/maps:
	// 559576822000-559576827000 r-xp 00002000 00:1a 4586   /usr/bin/cat

	return memProtectRX
}
