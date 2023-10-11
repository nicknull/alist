build:
	@gomobile bind -o ./AlistKit.xcframework -target=ios -iosversion=12.0 -ldflags="-s -w" ./lib
