package notes

/*
JSON 解析
	Go 自带函数库中的 encoding/json 库
		json.go
	json-iterator：条件编译方式
		jsoniter.go
		在运行或构建时加上 -tags=jsoniter 选项即可，例如：go run main.go -tags=jsoniter
条件编译：通过 +build 与 -tag 实现
	如：
	//go:build !jsoniter && !go_json && !(sonic && avx && (linux || windows || darwin) && amd64)
		当 tags 不为 jsoniter 时，使用该文件进行编译
	//go:build jsoniter
*/
