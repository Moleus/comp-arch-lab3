message: word: 'Hello, World!'
pointer: word: message
zero: word: 0
out_port: word: 1

start: nop
  loop: ld (pointer)
    out out_port
    jz end
    ld pointer
    inc
    st pointer
    jmp loop
  end: hlt
