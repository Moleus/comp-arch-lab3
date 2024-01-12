vector: word: interrupt
in_port: word: 0
out_port: word: 1
flag: word: 0
line_feed: word: 10

start: ei
spin_loop: ld flag
  jz spin_loop
  hlt

interrupt: in in_port
  out out_port
  cmp line_feed
  jnz returning
  ld flag
  inc
  st flag
  returning: iret
