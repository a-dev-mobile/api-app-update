package utils

import (
	"strings"
)

// GetActualStoreName translates user inputs to standardized store names, ensuring case insensitivity.
// The output store names are in lower case to maintain a consistent format that should match the
// lower-case store names stored in the database.
func GetActualStoreName(installerPackageName string) string {
	// Trim whitespace and convert to lower case for standardization
	storeTrimmed := strings.TrimSpace(installerPackageName)
	switch strings.ToLower(storeTrimmed) {

	case "ru.vk.store":
		return "rustore"

	case "com.huawei.appmarket", "com.huawei.localBackup":
		return "appgallery"

	case "com.android.vending":
		return "googleplay"

	case "com.apple":
		return "appstore"

	case "com.xiaomi.mipicks", "com.xiaomi.midrop":
		return "getapps"

	case "com.apple.testflight":
		return "testflight"

	case "com.apple.simulator":
		return "applesimulator"

	case "com.android.shell":
		return "adb"

	case "release.apk","apk","com.android.packageinstaller", "com.miui.packageinstaller", "com.google.android.packageinstaller":
		return "packageinstaller"

	case "com.sec.android.app.samsungapps", "com.samsung.android.scloud":
		return "samsungapps"

	// 'other' category for various installer package names that do not match the standard store names.
	case "com.lb.app_manager", "com.aefyr.sai", "ru.zdevs.zarchiver", "shareit.lite", "com.miui.huanji", "com.xrom.intl.appcenter", "com.huawei.gamebox", "com.hihonor.android.clone", "com.hicloud.android.clone", "com.vivo.easyshare", "com.coloros.backuprestore", "com.lenovo.anyshare.gps", "com.sec.android.easyMover", "com.apkpure.aegon", "com.aurora.store", "com.transsion.letswitch":
		return "other"
	default:
		// Return the original package name if it doesn't match any known store names.
		return installerPackageName
	}
}
