" place this in the init path (.vimrc)
au BufNewFile,BufRead *.ink set filetype=ink

" place this in $HOME/.vim/syntax/ink.vim
if exists("b:current_syntax")
    finish
endif

" ink syntax definition for vi/vim
syntax sync fromstart

" prefer hard tabs
set noexpandtab

" case
syntax match inkLabel "\v\:"
syntax match inkLabel "\v\-\>"
highlight link inkCase Label

" operators
syntax match inkOp "\v\~"
syntax match inkOp "\v\+"
syntax match inkOp "\v\-"
syntax match inkOp "\v\*"
syntax match inkOp "\v\/"
syntax match inkOp "\v\%"

syntax match inkOp "\v\&"
syntax match inkOp "\v\|"
syntax match inkOp "\v\^"

syntax match inkOp "\v\<"
syntax match inkOp "\v\>"
syntax match inkOp "\v\="
syntax match inkOp "\v\."
syntax match inkOp "\v\:\="
highlight link inkOp Operator

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

" builtin functions
syntax match builtinFunctionCall "\v[A-Za-z@!?][A-Za-z0-9@!?]*\(" contains=inkIdentifier,inkBuiltin
syntax keyword inkBuiltin load contained

syntax keyword inkBuiltin args contained
syntax keyword inkBuiltin in contained
syntax keyword inkBuiltin out contained
syntax keyword inkBuiltin dir contained
syntax keyword inkBuiltin make contained
syntax keyword inkBuiltin stat contained
syntax keyword inkBuiltin read contained
syntax keyword inkBuiltin write contained
syntax keyword inkBuiltin delete contained
syntax keyword inkBuiltin listen contained
syntax keyword inkBuiltin req contained
syntax keyword inkBuiltin rand contained
syntax keyword inkBuiltin urand contained
syntax keyword inkBuiltin time contained
syntax keyword inkBuiltin wait contained
syntax keyword inkBuiltin exec contained

syntax keyword inkBuiltin sin contained
syntax keyword inkBuiltin cos contained
syntax keyword inkBuiltin pow contained
syntax keyword inkBuiltin ln contained
syntax keyword inkBuiltin floor contained

syntax keyword inkBuiltin string contained
syntax keyword inkBuiltin number contained
syntax keyword inkBuiltin point contained
syntax keyword inkBuiltin char contained

syntax keyword inkBuiltin type contained
syntax keyword inkBuiltin len contained
syntax keyword inkBuiltin keys contained
highlight link inkBuiltin Keyword

" strings
syntax region inkString start=/\v'/ skip=/\v(\\.|\r|\n)/ end=/\v'/
highlight link inkString String

" comment
" -- block
syntax region inkComment start=/\v`/ skip=/\v(\\.|\r|\n)/ end=/\v`/ contains=inkTodo
highlight link inkComment Comment
" -- line-ending comment
syntax match inkLineComment "\v``.*" contains=inkTodo
highlight link inkLineComment Comment
" -- shebang, highlighted as comment
syntax match inkShebangComment "\v^#!.*"
highlight link inkShebangComment Comment
" -- TODO in comments
syntax match inkTodo "\v(TODO\(.*\)|TODO)" contained
syntax keyword inkTodo XXX contained
highlight link inkTodo Todo

" syntax-based code folds
syntax region inkExpressionList start="(" end=")" transparent fold
syntax region inkMatchExpression start="{" end="}" transparent fold
syntax region inkComposite start="\v\[" end="\v\]" transparent fold
set foldmethod=syntax

let b:current_syntax = "ink"
