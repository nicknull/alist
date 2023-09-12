build:
	@gomobile bind -o ./AlistKit.xcframework -target=ios -iosversion=11.0 -ldflags="-s -w" ./lib