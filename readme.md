# dig8

A DNS crawling library.

To install:

```
go get lonnie.io/dig8
```

## Source files (in reading order).

### Dns8: DNS crawling library

- `bug_on.go`: A helper for bug panic
- `printer.go`: A indented printer
- `check_label.go`: Checking label validity
- `pack_labels.go`: Packing labels into a DNS packet
- `unpack_labels.go`: Unpacking lables from a DNS packet
- `regmap.go`: Registrar name maps
- `domain.go`: Domain name
- `encoding.go`: Imports the big endian encoding
- `codes.go`: DNS packet field codes
- `flags.go`: DNS packet flag codes
- `question.go`: DNS query question data structure
- `rdata.go`: General rdata interface
- `rd_ipv4.go`: A records
- `rd_ipv6.go`: AAAA records
- `rd_domain.go`: NS, CNAME records
- `rd_mx.go`: MX records
- `rd_txt.go`: TXT records
- `rd_soa.go`: SOA records
- `rd_bytes.go`: Other records that we do not care
- `pack_rdata.go`: Rdata packing
- `unpack_rdata.go`: Rdata unpacking
- `ttl_str.go`: TTL string representation
- `rr.go`: A DNS record with domain and TTL
- `section.go`: Record section
- `selector.go`: A record selecter interface
- `sel_record.go`: Select domain record for of a particular type
- `sel_redirect.go`: Select redirection related record
- `sel_answer.go`: Select answer record
- `sel_ip.go`: Select IP address for a domain
- `packet.go`: DNS packet
- `dns_port.go`: DNS protocol port
- `message.go`: DNS message with a server
- `query.go`: DNS query message to a server
- `query_printer.go`: Query printer
- `err_timeout.go`: The timeout error
- `exchange.go`: A message exchange with a server
- `job.go`: A query job for a DNS client
- `id_pool.go`: A pool of DNS query ids, thread safe
- `client.go`: DNS client
- `node.go`: A query tree node
- `leaf.go`: A query tree leaf
- `branch.go`: Query tree branch, tasks and cursors
- `stack.go`: Stack for building a query tree
- `term_config.go`: Query terminal config
- `cursor.go`: Query terminal cursor
- `term.go`: Query terminal
- `ipuint.go`: Convert ip to uint32
- `name_server.go`: Name server for a zone
- `ns_shuffle.go`: Shuffling a list of name servers
- `zone_servers.go`: A set of name servers for a zone
- `cache_entry.go`: Name server cache entry
- `cache.go`: Name server cache
- `roots.go`: Root servers
- `recur.go`: Recursive query task
- `ips.go`: IP query task
- `info.go`: All record query task

### Dig8: DNS Crawler scheduler

- `task.go`: Crawling a single domain
- `respond.go`: Crawling respond
- `job.go`: Crawling a list of domains
- `job_request.go`: Job create request
- `worker.go`: Crawling worker