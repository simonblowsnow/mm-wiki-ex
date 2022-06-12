
go build -ldflags "-s -w -H=windowsgui" ./
taskkill /f /im mm-wiki.exe
start mm-wiki.exe --conf conf/mm-wiki.conf
