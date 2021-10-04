package main

const third string = `
: r 1 exit

: ]
  r @
  1 -
  r !
  _read
  ]

: main immediate ]

main

: '"'      32 exit
: ')'      41 exit
: '\n'     10 exit
: 'space'  32 exit
: '0'      48 exit
: '-'      45 exit

: cr '\n' echo exit

: _x  3 @ exit
: _x! 3 ! exit
: _y  4 @ exit
: _y! 4 ! exit

: swap _x! _y! _x _y exit

: +
  0 swap -
  -
  exit

: dup _x! _x _x exit

: h 0 exit

: inc
  dup @
  1 +
  swap
  ! exit

: ,
  h @
  !
  h inc
  exit

: '
  r @
  @
  dup
  1 +
  r @ !
  @
  exit

: ; immediate
  ' exit
  ,
  exit

: drop 0 * + ;

: dec dup @ 1 - swap ! ;


`

