hello: word: 12, 'Hello world!'
addr: word: hello
cnt: word: 0
out: word: 2047

start: ld hello
  st cnt
  loop: ld addr
      st addr
      ld cnt
      st cnt
      jz end
      jmp loop
  end: hlt