### Client update:
- client needs to look somewhat like server but with couple major changes:
    - need to solicit connections on port 2022 from localhost

#### client needs to:
- dial to server via a tcp, tcp+tls, quic conn
- implement an ErrorCollector function (or perhaps it should be an interace 🤔)
- sensibly call conn establishing funcs via go-routines