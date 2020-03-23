package subvert

func osGetMemoryProtection(address uintptr) memProtect {
	// TODO
	// https://stackoverflow.com/questions/1627998/retrieving-the-memory-map-of-its-own-process-in-os-x-10-5-10-6
	// https://stackoverflow.com/questions/9198385/on-os-x-how-do-you-find-out-the-current-memory-protection-level
	// https://www.grant.pizza/blog/using-dynamic-libraries-in-static-go-binaries/

	return memProtectRX
}
