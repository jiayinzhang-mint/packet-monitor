# packet-monitor
A BPF-based XDP Packet Filter

Example
'''
go build main.go
sudo ./main xdp load --ifindex=eth0 --verbose
'''
