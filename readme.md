# dig8

A DNS crawling library.

## Install by source

```
go get lonnie.io/dig8
```

## Source files.

- `bug_on.go`: A helper for bug panic.
- `check_label.go`: Checking label validity
- `pack.go`: Packing labels into a DNS packet.
- `unpack.go`: Unpacking lables from a DNS packet.
- `regmap.go`: Registrar name maps.
- `domain.go`: Domain name.
- `enc.go`: Imports the big endian encoding.
- `codes.go`: DNS packet field codes.
- `flags.go`: DNS packet flag codes.
- `question.go`: DNS query question data structure.
- `id_pool.go`: A pool of DNS query ids, thread safe.
