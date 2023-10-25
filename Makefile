build:
	@gomobile bind -o ./AlistKit.xcframework -target=ios,tvos -ldflags="-s -w" ./lib
