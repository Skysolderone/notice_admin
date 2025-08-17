build:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w"  -o notice  api/notice.go 
	scp notice wws:~
	rm notice
	ssh wws "mv notice zero_notice/notice && cd zero_notice && pkill notice && nohup ./notice &"