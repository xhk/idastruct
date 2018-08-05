package main
import (
	"fmt"
	"os"
	"io/ioutil"
	"strconv"
)

type Struct struct {
	Name string
	Members []Member

}

type Member struct{
	Name string
	TypeName string
	IsArr bool
	ArrLen int
}

type Word struct{
	Word string
	LineNo int
	ColNo int
}

func (this *Word) Equal(word string) bool{
	return this.Word == word
}

type StructParser struct{
	srcFile string
	Words []Word
	wordIndex int
	structs []Struct
}

func (this *StructParser) ParseFile(filePath string){
	fmt.Print("start parse ...\n")
	this.wordIndex = 0
	code, err := ioutil.ReadFile(filePath)
	if err != nil{
		if os.IsNotExist(err){
			fmt.Printf("%s not exist!\n", filePath)
			return
		}else{
			fmt.Println(err)
			return
		}
	}
	fmt.Printf("code:\n%s\n", string(code))

	fmt.Print("start split ...\n")
	this.splitWord(code)

	fmt.Print("start parse words...\n")
	this.parse();
}

func (this *StructParser) NextWord() (Word,bool) {
	if this.wordIndex >= len(this.Words){
		var word Word
		return word, false
	}

	return this.Words[this.wordIndex],true
}

func (this *StructParser) Rollback(){
	this.wordIndex --
}

func (this *StructParser) ScrollToNextLine() bool {
	for this.wordIndex>1 && this.wordIndex<len(this.Words)-1 && this.Words[this.wordIndex-1].LineNo == this.Words[this.wordIndex].LineNo {
		this.wordIndex++
	}

	return this.Words[len(this.Words)-1].LineNo != this.Words[this.wordIndex].LineNo
}

func (this *StructParser) parse(){
	var currStruct Struct
	w,ret := this.NextWord()
	for ret{
		if w.Equal("#pragma"){
			this.ScrollToNextLine()
		}else if w.Equal("struct"){
			w,ret = this.NextWord() 
			currStruct.Name = w.Word
		}else if w.Equal("{"){

		}else if w.Equal("}"){
			w,ret = this.NextWord() 
			if w.Equal(";"){
				this.structs = append(this.structs, currStruct)
				currStruct.Name = ""
				currStruct.Members = currStruct.Members[:0]
			}else{
				this.wordIndex--
			}
		}else{
			if len(currStruct.Name) !=0{
				var mem Member
				mem.TypeName = w.Word
				w,ret = this.NextWord()
				mem.Name = w.Word
				
				w,ret = this.NextWord()
				if w.Equal("["){
					mem.IsArr = true;
					w,ret = this.NextWord()
					mem.ArrLen,_ = strconv.Atoi(w.Word)
					w,ret = this.NextWord() // ']'
					w,ret = this.NextWord() // ';'
				}else{
					// ';' 不处理
				}
				currStruct.Members = append(currStruct.Members, mem)
			}
		}

		w,ret = this.NextWord()
	}
}

// 分词
func (this *StructParser) splitWord(code []byte) {
	var word Word
	words := make([]Word, 0)
	word.LineNo = 1
	for _,c := range code{
		if IsSeparator(c){
			if len(word.Word) > 0{
				words = append(words, word)
				word.Word = ""
			}
		}else if c =='\r' {

		}else if c == '\n'{
			if len(word.Word) > 0{
				words = append(words, word)
				word.Word = ""
			}
			word.LineNo ++ 
		}else if c == '{' || c == '}' || c =='(' || c ==')' || c=='[' || c==']' || c ==';'{
			if len(word.Word) > 0{
				words = append(words, word)
				word.Word = ""
			}
			word.Word += string(c)
			words = append(words, word)
			word.Word = ""
		}else if c == '/'{ // 这里没有注释可以先不处理

		}else {
			word.Word += string(c)
		}
	}
}

func IsSeparator(c byte) bool {
	return c ==' ' || c == '\t'
}

func (this *StructParser) PrintWords(){
	for _,w := range(this.Words){
		fmt.Printf("line:%d %s\n", w.LineNo, w.Word) 
	}
}
