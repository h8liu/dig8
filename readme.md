# dig8

A DNS crawling library.

To install:

```
go get lonnie.io/dig8
```

## Source files.

- `bug_on.go`: A helper for bug panic.
- `check_label.go`: Checking label validity
- `pack_labels.go`: Packing labels into a DNS packet.
- `unpack_labels.go`: Unpacking lables from a DNS packet.
- `regmap.go`: Registrar name maps.
- `domain.go`: Domain name.
- `encoding.go`: Imports the big endian encoding.
- `codes.go`: DNS packet field codes.
- `flags.go`: DNS packet flag codes.
- `question.go`: DNS query question data structure.
- `id_pool.go`: A pool of DNS query ids, thread safe.
- `rdata.go`: General rdata interface.
- `rdata_ipv4.go`: A records
- `rdata_ipv6.go`: AAAA records
- `rdata_domain.go`: NS, CNAME records
- `rdata_mx.go`: MX records
- `rdata_txt.go`: TXT records
- `rdata_soa.go`: SOA records
- `rdata_bytes.go`: Other records that we do not care
- `pack_rdata.go`: Rdata packing
- `unpack_rdata.go`: Rdata unpacking
- `rr.go`: A DNS record with domain and TTL.