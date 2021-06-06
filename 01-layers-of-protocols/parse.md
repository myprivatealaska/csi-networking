Parsing the **pcap-savefile**
https://www.tcpdump.org/manpages/pcap-savefile.5.txt

### Global header (24 octets)


1. **d4c3 b2a1** - _Magic number_ - allows software reading the file to determine
   whether the byte order of the host that wrote the file is the  same  as
   the  byte  order  of the host on which the file is being read, and thus
   whether the values in the per-file and per-packet headers  need  to  be
   byte-swapped.  The sender and the reader have different byte ordering 
2. **0200** - File format major version number - 2
3. **0400** - File format minor version number - 4
4. **0000 0000** - Timezone offset
5. **0000 0000** - Accuracy of timestamps in the file
6. **ea05 0000** - **Snapshot length** of the capture packets  longer  than  the  snapshot length are truncated to the
   snapshot length - 1514
7. **0100 0000** - Link-layer header type for packets in the capture - Ethernet

Following the per-file header are zero or  more  packets;  each  packet
begins  with  a per-packet header, which is immediately followed by the
raw packet data.  The format of the per-packet header is

### First packet header (16 octets)

4098 d057 0a1f 0300 4e00 0000 4e00 0000

1. **4098 d057** - Timestamp, seconds - 1473288256  Wed Sep 07 2016 22:44:16 GMT+0000
2. **0a1f 0300** - Timestamp, microseconds - time  in
   microseconds  or  nanoseconds since that second
3. **4e00 0000** - Length of captured packet data, bytes - 78
4. **4e00 0000** - Un-truncated length of the packet data, bytes - 78

