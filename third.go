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

: else immediate
  ' branch ,
  here
  0 ,
  swap
  dup here swap -
  swap !
;

: over _x! _y! _y _x _y ;

: add
  _x!
  _x @
  +
  _x !
;

: allot h add ;

: maybebranch
  logical
  r @ @ @
  computebranch
  r @ @ +
  r @ !
;

: mod _x! _y!
  _y _y _x / _x *
  -
;

: printnum
  dup
  10 mod '0' +
  swap 10 / dup
  if
    printnum
    echo
  else
    drop
    echo
  then
;

: .
  dup 0 <
  if
    '-' echo minus
  then
  printnum
  'space' echo
;

: debugprint dup . cr ;

: _print
  dup 1 +
  swap @
  dup '"' =
  if
    drop exit
  then
  echo
  tail _print
;

: print _print ;

: immprint
  r @ @
  print
  r @ !
;

: find-"
  key dup ,
  '"' =
  if
    exit
  then
  tail find-"
;

: " immediate
  key drop
  ' immprint ,
  find-"
;

: do immediate
  ' swap ,
  ' tor ,
  ' tor ,
  here
;

: i r @ 1 - @ ;
: j r @ 3 - @ ;

: > swap < ;
: <= 1 + < ;
: >= swap <= ;

: inci
  r @ 1 -
  inc
  r @ 1 - @
  r @ 2 - @
  <=
  if
    r @ @ @ r @ @ + r @ ! exit
  then
  fromr 1 +
  fromr drop
  fromr drop
  tor
;

: loop immediate ' inci @ here - , ;

: loopexit
  fromr drop
  fromr drop
  fromr drop
;

: execute
  8 !
  ' exit 9 !
  8 tor
;

: :: j

: fix-:: immediate 3 ' :: ! ;
fix-::

: : immediate :: ] ;

: command
  here 5 !
  _read
  here 5 @
  = if
    tail command
  then
  here 1 - h !
  here 5 @
  = if
    here @
    execute
  else
    here @
    here 1 - h !
  then
  tail command
;

: make-immediate
  here 1 -
  dup dup
  h !
  @ 
  swap
  1 -
  !
;

: <build immediate
  make-immediate
  ' :: ,
  -1 , ( compile 'pushint' was 2 in original )
  here 0 ,
  ' , ,
;

: does> immediate
  ' command ,
  here swap !
  -2 , (compile run-code primitive, was 2 in original )
  ' fromr ,
;

: _dump
  dup " (" . " , "
  dup @
  dup ' exit
  = if
    " ;)" cr exit
  then
  . " ), "
  1 +
  tail _dump
;

: dump _dump ;

: # . cr ;

: var <build , does> ;
: constant <build , does> @ ;
: array <build allot does> + ;

: [ immediate command ;
: _welcome " Welcome to THIRD.
Ok.
" ;

: ; immdeiate ' exit , command exit

[

_welcome
`
