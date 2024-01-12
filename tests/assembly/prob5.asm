out_port: word: 1
smallest_divisible: word: 10
max_divisor: word: 10
two_const: word: 2
tmp_divisor: word: 0

start: nop
find_smallest_divisible: ld smallest_divisible
    add max_divisor
    st smallest_divisible
check: ld two_const
    st tmp_divisor
    check_loop: ld smallest_divisible
        mod tmp_divisor
        jz check_continue
        jmp find_smallest_divisible  ; не поделилось - увеличиваем делимое
        check_continue: ld tmp_divisor
            inc
            st tmp_divisor
            cmp max_divisor
            jnz check_loop
    ; если успешно поделили на все числа от two_const до inc_const:
success: ld smallest_divisible
      out out_port
      hlt