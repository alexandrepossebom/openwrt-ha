# openwrt-ha
Integration with openwrt and HomeAssistant

This simple integration will add two entities in HA:

sensor.wan_tx
sensor.wan_rx

Traffic in MBps

To use you can download the binary in releases or you need to cross compile it to run in openwrt mips:

```
GOOS=linux GOARCH=mips GOMIPS=softfloat go build -ldflags="-s -w"
```

upload you new binary to you openwrt and run:

```
./openwrt-ha -token eyJ0eXAi... \
    -ha 192.168.1.100 \
    -wan eth0.1
```

If you prefer, you can place the token in the /etc/config/ha-token file.

