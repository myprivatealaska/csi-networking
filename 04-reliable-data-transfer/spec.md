### Reliable data delivery ###

Scenario 1.
Bits in the packet might be corrupted. 
Let's implement an Automatic Repeat reQuest protocol with 
 - error detection (based off of the checksum)
 - receiver feedback (ACK + NAK)
 - retransmission

#### Header ####
uint16 source port # \
uint16 dest port # \
uint16 checksum



