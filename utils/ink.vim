" place this in the init path (.vimrc)
au BufNewFile,BufRead *.ink set filetype=ink

" place this in $HOME/.vim/syntax/ink.vim
"  (I have it symlinked)
if exists("b:current_syntax")
    finish
endif

" ink syntax highlight definition for vi/vim

" delimiters ()[]{}
syntax match inkDelim "\v[\(\)\[\]\{\}]"
highlight link inkDelim Delimiter

" operators
syntax match inkOp "\v\~"
syntax match inkOp "\v\+"
syntax match inkOp "\v\-"
syntax match inkOp "\v\*"
syntax match inkOp "\v\/"
syntax match inkOp "\v\%"

syntax match inkOp "\v\<"
syntax match inkOp "\v\>"
syntax match inkOp "\v\="
syntax match inkOp "\v\."
syntax keyword inkOp is
highlight link inkOp Operator

" case
syntax match inkLabel "\v\:"
syntax match inkLabel "\v\:\="
syntax match inkLabel "\v\-\>"
highlight link inkCase Label

" match
syntax match inkMatch "\v\:\:"
highlight link inkMatch Conditional

" functions
syntax match inkFunction "\v\=\>"
highlight link inkFunction Function

" booleans
syntax keyword inkBoolean true false
highlight link inkBoolean Boolean

" numbers should be consumed first by identifiers, so comes before
syntax match inkNumber "\v\d+"
syntax match inkNumber "\v\d+\.\d+"
highlight link inkNumber Number

" identifiers
syntax match inkIdentifier "\v[A-Za-z@!?][A-Za-z0-9@!?]*"
syntax match inkIdentifier "\v_"
highlight link inkIdentifier Identifier

" strings
syntax region inkString start=/\v'/ skip=/\v(\\.|\r|\n)/ end=/\v'/
highlight link inkString String

" comment
syntax region inkComment start=/\v`/ skip=/\v(\\.|\r|\n)/ end=/\v`/
highlight link inkComment Comment

let b:current_syntax = "ink"
