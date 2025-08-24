build:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w"  -o notice  api/notice.go 
	scp notice wws:~
	scp etc/api.yaml wws:~/zero_notice/etc/
	rm notice
	ssh wws "mv notice zero_notice/notice && cd zero_notice && pkill notice" 
	ssh wws