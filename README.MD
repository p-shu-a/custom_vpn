UDP vs TCP
- differences?
- uses?

- what is NORDVPN's architecture?

- to start sshd on macos, you gotta enable remote login. do this via settings panel or :
sudo systemsetup -setremotelogin on

- how does go handle go-routines which spawn go-routines?

user@user ~ $ lsof -Pn -i6 | grep "^server\|^client"
server    30645 user    3u  IPv6 0xf18cdf112a105407      0t0  TCP *:9000 (LISTEN)
server    30645 user    5u  IPv6 0x677e1030533d680b      0t0  TCP [::1]:9000->[::1]:54632 (ESTABLISHED)
server    30645 user    6u  IPv6 0x489670b6cc14f73b      0t0  TCP [::1]:54633->[::1]:22 (ESTABLISHED)
client    30646 user    3u  IPv6 0xc37b1a3b2f44f182      0t0  TCP *:2022 (LISTEN)
client    30646 user    5u  IPv6 0x5a5707c7b5ce2824      0t0  TCP [::1]:2022->[::1]:54631 (ESTABLISHED)
client    30646 user    6u  IPv6 0xec586c6d9524a865      0t0  TCP [::1]:54632->[::1]:9000 (ESTABLISHED)