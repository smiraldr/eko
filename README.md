# eko
CLI app built using golang for pinging ipv4 addresses Library used: Cobra net (for icmp and ipv4)

Usage: go install eko (in the folder)
Works only as root due to web socket permissions, to run as admin : sudo -s 

on terminal eko --help for further usage.

example ping
eko pingip4 -4 google.com

Adding ipv6 support soon ! Check back
