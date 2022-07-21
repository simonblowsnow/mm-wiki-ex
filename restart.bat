
go build -ldflags "-s -w -H=windowsgui" ./
taskkill /f /im mm-wiki-ex.exe
start mm-wiki-ex.exe --conf conf/mm-wiki.conf
