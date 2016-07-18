
syntax clear

syn match comment "^[^\n#]*#.*$"
syn match imports " as .*$" 
syn match colon "\:\|\-"
syn match lingoLexiconBranch    "tenets\:\|lexicons\:"  contains=colon
syn match lingoDoc2    "\(comment\)\:.*$" contains=regstring,colon
syn match docstring       "^\(\t\| \)*doc\:.*$"
syn region  regstring          oneline start='"' end='"' containedin=lingoDoc2
syn match nameTypeLit "[a-zA-Z][a-zA-Z0-9\-]*$" containedin=nameLit
syn match typeLit "[a-zA-Z][a-zA-Z0-9\-]*$"
syn match deRef "\$[a-zA-Z][a-zA-Z0-9\-]*$"
syn match nameLit "- *name\: *.*$" contains=name,nameTypeLit,colon
syn match pipeback "^\(\t\| \)*\(<\|\!\)" nextgroup=lingolabel
syn match lingoLab "[a-z\-_]\+\(\[[0-9]+\]\)\?:" nextgroup=expr
syn match expr "[0-9\=\>\< ]\+$"

hi link lingoLexiconBranch Label 
hi link lingoLab Type	 
hi link nameLit Special
hi link deRef Operator	 
hi link typeLit Special 
hi link nameTypeLit TypeDef
hi link nameTypeLit Underlined
hi link lingoDoc Define
hi link regstring Statement 
hi link docstring Comment
hi link pipeback Special 
hi link expr Special
hi link colon Special 
hi link imports Special 
hi link comment Comment
