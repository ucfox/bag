" --------------------------------
" 项目级 Vim 配置文件

if exists("pylon_prj_loaded")
    finish
endif

let pylon_prj_loaded = 1

" --------------------------------

" 获取项目根目录
let s:prjroot=fnamemodify('',':p')

" 解除默认的单元测试映射
unmap <F2>

" 定义单元测试的映射
noremap <F2> <Esc> :call MapPrjUnitTest() <CR>

" 定义单元测试函数
" function MapPrjUnitTest()
"     if filereadable(s:prjroot.'test/unittest.sh')
"         exec '!'. s:prjroot .'test/unittest.sh'
"     else
"         exec '! /home/q/tools/pylon_rigger/rigger start -s test'
"     endif
" endfunction
function MapPrjUnitTest()
    if filereadable(s:prjroot.'test/gotest.sh')
        exec '!'. s:prjroot .'test/gotest.sh'
    else
        exec '! /home/q/tools/pylon_rigger/rigger start -s test'
    endif
endfunction

" 定义golang ctags的映射
noremap <F6> <Esc> :call MapGolangTags() <CR>
" 定义golang ctags 生成函数
function MapGolangTags()
    exec '! /home/q/tools/game_team/bin/gotags -tag-relative=true -f ./_prj/golang_tags -R src/*'
    exec '! //home/q/tools/game_team/bin/gotags -tag-relative=false -f ./_prj/golang_src_tags -R /usr/local/go/src/'
    :call UpdatePrjTags()
endfunction

func! UpdatePrjTags()
    " 将 _prj/ 下 tags 结尾的文件加入tags
    let a:tag_list=split(globpath(s:prjroot."_prj/", '*tags'), "\n")
    set tags=
    let i=0
    while i<len(a:tag_list)
        if filereadable(a:tag_list[i])
            " echo a:tag_list[i]
            execute "set tags+=".a:tag_list[i]
        endif
        let i+=1
    endwhile
endf
