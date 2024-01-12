vector: word: interrupt
message_pointer: word: message
message: word: 'What is your name?'
greeting_pointer: word: greeting
greeting: word: 'Hello, '
exclamation_pointer: word: exclamation
exclamation: word: '!'
in_port: word: 0
out_port: word: 1
flag: word: 0
line_feed: word: 10

start: nop
    message_loop: ld (message_pointer)
        jz message_loop_end
        out out_port
        ld message_pointer
        inc
        st message_pointer
        jmp message_loop
    ; ожидание ввода
    message_loop_end: ld line_feed
    out out_port
    ei
    spin_loop: ld flag
        jz spin_loop
    ; вывод приветствия
    di
    greeting_loop: ld (greeting_pointer)
        jz name_loop
        out out_port
        ld greeting_pointer
        inc
        st greeting_pointer
        jmp greeting_loop
    name_loop: ld (buffer_start_pointer)
        jz exclamation_printing
        out out_port
        ld buffer_start_pointer
        inc
        st buffer_start_pointer
        jmp name_loop
    exclamation_printing: ld (exclamation_pointer)
        out out_port
    hlt

interrupt: in in_port ; получение слова из порта ввода
    cmp line_feed
    jnz continue
    ld flag
    inc
    st flag
    jmp returning
    continue: st (buffer_pointer)
        ld buffer_pointer
        inc
        st buffer_pointer
    returning: iret

buffer_start_pointer: word: buffer
buffer_pointer: word: buffer
buffer: word: 0
