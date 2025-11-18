package programs

import (
	"golang.org/x/sys/windows/registry"
)

func FindSteam() (string, error) {
	// 尝试从 64 位注册表读取
	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\WOW6432Node\Valve\Steam`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		// 尝试从 32 位注册表读取
		key, err = registry.OpenKey(
			registry.LOCAL_MACHINE,
			`SOFTWARE\Valve\Steam`,
			registry.QUERY_VALUE,
		)
		if err != nil {
			return "", err
		}
	}
	defer key.Close()

	path, _, err := key.GetStringValue("InstallPath")
	if err != nil {
		return "", err
	}
	return path, nil
}
