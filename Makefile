build:
	@gomobile bind -o ./AlistKit.xcframework -target=ios -iosversion=15.0 -ldflags="-s -w" ./lib