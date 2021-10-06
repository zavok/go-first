package main

var third string = `

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

: tor
  r @ @
  swap
  r @ !
  r @ 1 + r !
  r @ !
;

: fromr
  r @ @
  r @ 1 - r !
  r @ @
  swap
  r @ !
;

: tail fromr fromr drop tor ;

: minus 0 swap - ;

: bnot 1 swap - ;

: < - <0 ;

: logical
  dup
  0 <
  swap minus
  0 <
  +
;

: not logical bnot ;

: = - not ;

: branch
  r @
  @
  @
  r @ @
  +
  r + !
;

: computebranch 1 - * 1 + ;

: notbranch
  not
  r @ @ @
  computebranch
  r @ @ +
  r @ !
;

: here h @ ;

: if immediate
  ' notbranch ,
  here
  0 ,
;

: then immediate
  dup
  here
  swap -
  swap !
;

: find-)
  key
  ')' =
  not if
    tail find-)
  then
;

: ( immediate
  find-)
;

( we should be able to do FORTH-style comments now )

( now that we've got comments, we can comment the rest of the code
  in a legitimate [self parsing] fashion. Note that you can't
  nest parentheses... )

`
