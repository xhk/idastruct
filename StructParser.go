package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

type Struct struct {
	Name    string
	Members []Member
}

func (this *Struct) ToString() string {
	var s = ""
	s += fmt.Sprintf("struct %s\n", this.Name)
	s += "{\n"
	for _,m := range this.Members{
		s += "\t" + m.ToString() + "\n"
	}

	s += "};\n"

	return s
}

type Member struct {
	Name     string
	TypeName string
	IsArr    bool
	ArrLen   int
}

func (this *Member) ToString() string {
	if this.IsArr {
		return fmt.Sprintf("%s %s[%d];", this.TypeName, this.Name, this.ArrLen)
	}
	return fmt.Sprintf("%s %s;", this.TypeName, this.Name)
}

// 内置类型的大小
var inner_types = map[string]int{
	"int":                4,
	"char":               1,
	"CString":            4,
	"CMapStringToString": 0x1C,
}

// user define types
var user_types = map[string]int{}

func (this *Member) Size() int {
	if size, ok := inner_types[this.TypeName]; ok {
		if this.IsArr{
			return size * this.ArrLen
		}
		return size
	}

	if size, ok := user_types[this.TypeName]; ok {
		if this.IsArr{
			return size * this.ArrLen
		}
		return size
	}

	return 0
}

func (this *Member) Offset() int {
	ret := 0
	pos := 0
	for i, c := range this.Name {
		pos = i
		if c >='0' && c <= '9' {
			break
		}
	}

	ret,_ = strconv.Atoi(this.Name[pos:])
	return ret
}

type Word struct {
	Word   string
	LineNo int
	ColNo  int
}

func (this *Word) Equal(word string) bool {
	return this.Word == word
}

type StructParser struct {
	srcFile   string
	Words     []Word
	wordIndex int
	structs   []Struct
	fixedStructs [] Struct // 修正后的
}

func (this *StructParser) ParseFile(filePath string) {
	fmt.Print("start parse ...\n")
	this.wordIndex = 0
	code, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("%s not exist!\n", filePath)
			return
		} else {
			fmt.Println(err)
			return
		}
	}
	fmt.Printf("code:\n%s\n", string(code))

	fmt.Print("start split ...\n")
	this.splitWord(code)
	//this.PrintWords()
	fmt.Print("start parse words...\n")
	this.parse()
	//this.DumpSrcStructs()
	this.Fix()
}

func (this *StructParser) NextWord() (Word, bool) {
	//fmt.Printf("wordIndex:%d\n", this.wordIndex)
	if this.wordIndex >= len(this.Words) {
		var word Word
		return word, false
	}

	w := this.Words[this.wordIndex]
	this.wordIndex++
	return w, true
}

func (this *StructParser) Rollback() {
	this.wordIndex--
}

func (this *StructParser) ScrollToNextLine() bool {
	for this.wordIndex > 1 && this.wordIndex < len(this.Words)-1 && this.Words[this.wordIndex-1].LineNo == this.Words[this.wordIndex].LineNo {
		this.wordIndex++
	}

	return this.Words[len(this.Words)-1].LineNo != this.Words[this.wordIndex].LineNo
}

func (this *StructParser) parse() {
	var currStruct Struct
	w, ret := this.NextWord()
	for ret {
		//fmt.Printf("word:%s\n", w.Word)
		if w.Equal("#pragma") {
			this.ScrollToNextLine()
		} else if w.Equal("struct") {
			w, ret = this.NextWord()
			currStruct.Name = w.Word
		} else if w.Equal("{") {

		} else if w.Equal("}") {
			w, ret = this.NextWord()
			if w.Equal(";") {
				this.structs = append(this.structs, currStruct)
				currStruct.Name = ""
				currStruct.Members = currStruct.Members[:0]
			} else {
				this.wordIndex--
			}
		} else {
			if len(currStruct.Name) != 0 {
				var mem Member
				mem.TypeName = w.Word
				w, ret = this.NextWord()
				mem.Name = w.Word

				w, ret = this.NextWord()
				if w.Equal("[") {
					mem.IsArr = true
					w, ret = this.NextWord()
					mem.ArrLen, _ = strconv.Atoi(w.Word)
					w, ret = this.NextWord() // ']'
					w, ret = this.NextWord() // ';'
				} else {
					// ';' 不处理
				}
				currStruct.Members = append(currStruct.Members, mem)
			}
		}

		w, ret = this.NextWord()
	}
}

// 分词
func (this *StructParser) splitWord(code []byte) {
	var word Word
	words := make([]Word, 0)
	word.LineNo = 1
	for _, c := range code {
		if IsSeparator(c) {
			if len(word.Word) > 0 {
				words = append(words, word)
				word.Word = ""
			}
		} else if c == '\r' {

		} else if c == '\n' {
			if len(word.Word) > 0 {
				words = append(words, word)
				word.Word = ""
			}
			word.LineNo++
		} else if c == '{' || c == '}' || c == '(' || c == ')' || c == '[' || c == ']' || c == ';' {
			if len(word.Word) > 0 {
				words = append(words, word)
				word.Word = ""
			}
			word.Word += string(c)
			words = append(words, word)
			word.Word = ""
		} else if c == '/' { // 这里没有注释可以先不处理

		} else {
			word.Word += string(c)
		}
	}
	this.Words = words
}

func IsSeparator(c byte) bool {
	return c == ' ' || c == '\t'
}

func (this *StructParser) PrintWords() {
	for _, w := range this.Words {
		fmt.Printf("line:%d %s\n", w.LineNo, w.Word)
	}
}

func (this *StructParser) Fix() {
	for _, s := range this.structs {
		this.fixedStructs = append(this.fixedStructs, this.FixStruct(&s))
	}
}

func (this *StructParser) DumpFixedStructs(){
	for _, s := range this.fixedStructs{
		fmt.Print(s.ToString())
	}
}

func (this *StructParser) DumpSrcStructs(){
	for _, s := range this.structs{
		fmt.Print(s.ToString())
	}
}

func (this *StructParser) MemberIndex(memName string) int {
	ret := 0
	pos := 0
	for i, c := range memName {
		pos = i
		if c >='0' && c <= '9' {
			break
		}
	}

	ret,_ = strconv.Atoi(memName[pos:])
	return ret
}


func genMember(s * Struct, beginIndex int, size int){
	var count = 0
	var newMem Member

	if size % 4 == 0{ // 可以用整型
		count = size/4
		for i:=0;i<count;i++{ // 整型就不用数组了
			
			newMem = Member{
				fmt.Sprintf( "n%d", beginIndex + i * 4),
				"int",
				false,
				0,
			}
			s.Members = append(s.Members, newMem)
		}
	}else {
		if size>1{// 用 char []
			newMem = Member{
				fmt.Sprintf( "n%d", beginIndex),
				"char",
				true,
				size,
			}
		}else{
			newMem = Member{
				fmt.Sprintf( "n%d", beginIndex),
				"char",
				false,
				0,
			}
		}
		s.Members = append(s.Members, newMem)
	}
}

func (this *StructParser) FixStruct(s *Struct) Struct {
	var ret Struct
	ret.Name = s.Name
	pos := 0
	offset := 0
	for i, m := range s.Members {
		pos = m.Offset()
		if i == 0 {
			if pos != 0 {
				genMember(&ret, 0, pos-0)
				offset += pos-0
			} 

			ret.Members = append(ret.Members, m)
			offset += m.Size()
		} else {
			var div = pos - offset
			if div > 0{ // 有空隙
				genMember(&ret, offset, div)
				offset += div
			}

			ret.Members = append(ret.Members, m)
			offset += m.Size()
		}
	}

	return ret
}
