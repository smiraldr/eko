# eko
CLI app built using golang for pinging ipv4 and ipv6 addresses Library used: Cobra, net (for icmp,ipv4,ipv6)

Usage: go install eko (in the folder)
Works only as root due to web socket permissions, to run as admin : sudo -s 

On terminal eko --help for further usage.

example ping ipv4
eko pingip4 -4 google.com (or any ip4 address)

example ping ipv6
eko pingip6 -6 google.com (or any ip6 address)
