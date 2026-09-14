#!/bin/sh

set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
project_dir=$(CDPATH= cd -- "$script_dir/.." && pwd)

android_sdk=${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}
if [ -z "$android_sdk" ]; then
	for candidate in /opt/homebrew/share/android-commandlinetools "$HOME/Library/Android/sdk"; do
		if [ -x "$candidate/platform-tools/adb" ]; then
			android_sdk=$candidate
			break
		fi
	done
fi
if [ -z "$android_sdk" ] || [ ! -x "$android_sdk/platform-tools/adb" ]; then
	printf '%s\n' "Android SDK not found. Set ANDROID_HOME or ANDROID_SDK_ROOT." >&2
	exit 1
fi

if [ -z "${JAVA_HOME:-}" ]; then
	for candidate in \
		/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home \
		/Applications/Android\ Studio.app/Contents/jbr/Contents/Home; do
		if [ -x "$candidate/bin/java" ]; then
			JAVA_HOME=$candidate
			break
		fi
	done
fi
if [ -z "${JAVA_HOME:-}" ] || [ ! -x "$JAVA_HOME/bin/java" ]; then
	printf '%s\n' "JDK 17 not found. Set JAVA_HOME." >&2
	exit 1
fi

export ANDROID_HOME=$android_sdk
export ANDROID_SDK_ROOT=$android_sdk
export JAVA_HOME
PATH=$JAVA_HOME/bin:$android_sdk/platform-tools:$PATH
export PATH

ndk_dir=$android_sdk/ndk/28.2.13676358
if [ ! -d "$ndk_dir" ]; then
	printf '%s\n' "Android NDK 28.2.13676358 not found under $android_sdk/ndk." >&2
	exit 1
fi
ANDROID_NDK_HOME=$ndk_dir
export ANDROID_NDK_HOME

mkdir -p "$project_dir/android/app/libs"

cd "$project_dir"
go run github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@v2.9.11 bind \
	-target android/arm64 \
	-androidapi 23 \
	-javapkg com.olivierh59500 \
	-trimpath \
	-ldflags="-s -w" \
	-o android/app/libs/vectorballs.aar \
	./mobile

"$project_dir/android/gradlew" \
	--no-daemon \
	-p "$project_dir/android" \
	:app:assembleDebug

adb=$android_sdk/platform-tools/adb
if [ -z "${ANDROID_SERIAL:-}" ]; then
	device_count=$($adb devices | awk '$2 == "device" { count++ } END { print count + 0 }')
	if [ "$device_count" -ne 1 ]; then
		printf '%s\n' "Expected one authorized Android device; set ANDROID_SERIAL when several are connected." >&2
		exit 1
	fi
	ANDROID_SERIAL=$($adb devices | awk '$2 == "device" { print $1; exit }')
fi

"$adb" -s "$ANDROID_SERIAL" install -r "$project_dir/android/app/build/outputs/apk/debug/app-debug.apk"
"$adb" -s "$ANDROID_SERIAL" shell am force-stop com.olivierh59500.vectorballs
"$adb" -s "$ANDROID_SERIAL" shell am start -W -n com.olivierh59500.vectorballs/.MainActivity
