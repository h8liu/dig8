# dig8

A DNS crawling library.

To install:

```
go get github.com/h8liu/dig8
```

## To batch crawl a list of domains

```
$ dig8s <domain list file>
```

See `example` folder for usage and example output files.

## To crawl a single domain

```
$ dig8 lonnie.io
// lonnie.io
info lonnie.io {
    ips lonnie.io {
        // zone: .
        lonnie.io a @h.root-servers.net(128.63.2.53) {
            #24565 
            ques lonnie.io a
            auth {
                io ns a.nic.io 2d
                io ns a.ns13.net 2d
                io ns b.nic.ac 2d
                io ns b.nic.io 2d
                io ns b.ns13.net 2d
                io ns ns1.communitydns.net 2d
                io ns ns3.icb.co.uk 2d
            }
            addi {
                a.nic.io a 64.251.31.179 2d
                a.ns13.net a 49.212.31.192 2d
                b.nic.ac a 78.104.145.37 2d
                b.nic.io a 194.0.2.1 2d
                b.ns13.net a 49.212.51.85 2d
                ns1.communitydns.net a 194.0.1.1 2d
                ns3.icb.co.uk a 91.208.95.130 2d
                b.nic.io aaaa 2001:678:5::1 2d
                ns1.communitydns.net aaaa 2001:678:4::1 2d
            }
            (in 71.49ms)
        }
        // zone: io
        lonnie.io a @b.ns13.net(49.212.51.85) {
            #2085 
            ques lonnie.io a
            auth {
                lonnie.io ns dns4.registrar-servers.com 1d
                lonnie.io ns dns1.registrar-servers.com 1d
                lonnie.io ns dns3.registrar-servers.com 1d
                lonnie.io ns dns2.registrar-servers.com 1d
                lonnie.io ns dns5.registrar-servers.com 1d
            }
            (in 114.41ms)
        }
        // zone: lonnie.io
        ips dns1.registrar-servers.com {
            // zone: .
            dns1.registrar-servers.com a @b.root-servers.net(192.228.79.201) {
                #49138 
                ques dns1.registrar-servers.com a
                auth {
                    com ns l.gtld-servers.net 2d
                    com ns c.gtld-servers.net 2d
                    com ns j.gtld-servers.net 2d
                    com ns h.gtld-servers.net 2d
                    com ns a.gtld-servers.net 2d
                    com ns i.gtld-servers.net 2d
                    com ns b.gtld-servers.net 2d
                    com ns e.gtld-servers.net 2d
                    com ns d.gtld-servers.net 2d
                    com ns g.gtld-servers.net 2d
                    com ns k.gtld-servers.net 2d
                    com ns f.gtld-servers.net 2d
                    com ns m.gtld-servers.net 2d
                }
                addi {
                    a.gtld-servers.net a 192.5.6.30 2d
                    b.gtld-servers.net a 192.33.14.30 2d
                    c.gtld-servers.net a 192.26.92.30 2d
                    d.gtld-servers.net a 192.31.80.30 2d
                    e.gtld-servers.net a 192.12.94.30 2d
                    f.gtld-servers.net a 192.35.51.30 2d
                    g.gtld-servers.net a 192.42.93.30 2d
                    h.gtld-servers.net a 192.54.112.30 2d
                    i.gtld-servers.net a 192.43.172.30 2d
                    j.gtld-servers.net a 192.48.79.30 2d
                    k.gtld-servers.net a 192.52.178.30 2d
                    l.gtld-servers.net a 192.41.162.30 2d
                    m.gtld-servers.net a 192.55.83.30 2d
                    a.gtld-servers.net aaaa 2001:503:a83e::2:30 2d
                }
                (in 15.45ms)
            }
            // zone: com
            dns1.registrar-servers.com a @h.gtld-servers.net(192.54.112.30) {
                #31317 
                ques dns1.registrar-servers.com a
                auth {
                    registrar-servers.com ns dns1.name-services.com 2d
                    registrar-servers.com ns dns2.name-services.com 2d
                    registrar-servers.com ns dns3.name-services.com 2d
                    registrar-servers.com ns dns4.name-services.com 2d
                    registrar-servers.com ns dns5.name-services.com 2d
                }
                addi {
                    dns1.name-services.com a 98.124.192.1 2d
                    dns2.name-services.com a 98.124.197.1 2d
                    dns3.name-services.com a 98.124.193.1 2d
                    dns4.name-services.com a 98.124.194.1 2d
                    dns5.name-services.com a 98.124.196.1 2d
                }
                (in 143.50ms)
            }
            // zone: registrar-servers.com
            dns1.registrar-servers.com a @dns1.name-services.com(98.124.192.1) {
                #12790 auth
                ques dns1.registrar-servers.com a
                answ dns1.registrar-servers.com a 216.87.155.33 30m
                (in 68.69ms)
            }
            // dns1.registrar-servers.com(216.87.155.33)
        }
        lonnie.io a @dns1.registrar-servers.com(216.87.155.33) {
            #25537 auth
            ques lonnie.io a
            answ lonnie.io a 66.147.240.181 30m
            auth {
                lonnie.io ns dns1.registrar-servers.com 30m
                lonnie.io ns dns3.registrar-servers.com 30m
                lonnie.io ns dns5.registrar-servers.com 30m
                lonnie.io ns dns2.registrar-servers.com 30m
                lonnie.io ns dns4.registrar-servers.com 30m
            }
            (in 8.43ms)
        }
    }
    recur lonnie.io ns {
        // zone: lonnie.io
        lonnie.io ns @dns1.registrar-servers.com(216.87.155.33) {
            #43039 auth
            ques lonnie.io ns
            answ {
                lonnie.io ns dns1.registrar-servers.com 30m
                lonnie.io ns dns3.registrar-servers.com 30m
                lonnie.io ns dns2.registrar-servers.com 30m
                lonnie.io ns dns4.registrar-servers.com 30m
                lonnie.io ns dns5.registrar-servers.com 30m
            }
            (in 8.39ms)
        }
    }
    recur lonnie.io mx {
        // zone: lonnie.io
        lonnie.io mx @dns1.registrar-servers.com(216.87.155.33) {
            #55563 auth
            ques lonnie.io mx
            auth lonnie.io soa dns1.registrar-servers.com/hostmaster.registrar-servers.com serial=2015070800 refresh=43200 retry=3600 exp=604800 min=3601 1h1s
            (in 4.94ms)
        }
        // record does not exist
    }
    recur lonnie.io soa {
        // zone: lonnie.io
        lonnie.io soa @dns1.registrar-servers.com(216.87.155.33) {
            #59920 auth
            ques lonnie.io soa
            answ lonnie.io soa dns1.registrar-servers.com/hostmaster.registrar-servers.com serial=2015070800 refresh=43200 retry=3600 exp=604800 min=3601 1h1s
            auth {
                lonnie.io ns dns1.registrar-servers.com 30m
                lonnie.io ns dns4.registrar-servers.com 30m
                lonnie.io ns dns2.registrar-servers.com 30m
                lonnie.io ns dns5.registrar-servers.com 30m
                lonnie.io ns dns3.registrar-servers.com 30m
            }
            (in 4.54ms)
        }
    }
    recur lonnie.io txt {
        // zone: lonnie.io
        lonnie.io txt @dns1.registrar-servers.com(216.87.155.33) {
            #48545 auth
            ques lonnie.io txt
            auth lonnie.io soa dns1.registrar-servers.com/hostmaster.registrar-servers.com serial=2015070800 refresh=43200 retry=3600 exp=604800 min=3601 1h1s
            (in 5.31ms)
        }
        // record does not exist
    }
    // lonnie.io(66.147.240.181)
    
    // lonnie.io ns dns1.registrar-servers.com(216.87.155.33)
    
    // lonnie.io a 66.147.240.181
    // lonnie.io ns dns4.registrar-servers.com
    // lonnie.io ns dns1.registrar-servers.com
    // lonnie.io ns dns3.registrar-servers.com
    // lonnie.io ns dns2.registrar-servers.com
    // lonnie.io ns dns5.registrar-servers.com
    // dns1.registrar-servers.com a 216.87.155.33
    // lonnie.io soa dns1.registrar-servers.com/hostmaster.registrar-servers.com serial=2015070800 refresh=43200 retry=3600 exp=604800 min=3601
}
ips {
    66.147.240.181
}
servers {
    lonnie.io ns dns1.registrar-servers.com(216.87.155.33)
}
records {
    lonnie.io a 66.147.240.181
    lonnie.io ns dns4.registrar-servers.com
    lonnie.io ns dns1.registrar-servers.com
    lonnie.io ns dns3.registrar-servers.com
    lonnie.io ns dns2.registrar-servers.com
    lonnie.io ns dns5.registrar-servers.com
    dns1.registrar-servers.com a 216.87.155.33
    lonnie.io soa dns1.registrar-servers.com/hostmaster.registrar-servers.com serial=2015070800 refresh=43200 retry=3600 exp=604800 min=3601
}
```
