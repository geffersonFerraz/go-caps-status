CGO_ENABLED=1 go build -tags "embed" -ldflags "-s -w" -o gocaps

#now go to: https://askubuntu.com/questions/48321/how-do-i-start-applications-automatically-on-login