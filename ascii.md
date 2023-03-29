RFC8259:
空格： '\ ', '\t', '\r', '\n' 

JSON格式是1999年《JavaScript Programming Language, Standard ECMA-262 3rd Edition》的子集合，所以可以在JavaScript以eval()函数（javascript通过eval()调用解析器）读入。不过这并不代表JSON无法使用于其他语言，事实上几乎所有与网络开发相关的语言都有JSON函数库。

JSON的基本数据类型： false / null / true / object / array / number / string
number = [ minus ] int [ frac ] [ exp ]

数值：十进制数，不能有前导0，可以为负数，可以有小数部分。还可以用e或者E表示指数部分。不能包含非数，如NaN。不区分整数与浮点数。JavaScript用双精度浮点数表示所有数值。
字符串：以双引号""括起来的零个或多个Unicode码位。支持反斜杠开始的转义字符序列。
布尔值：表示为true或者false。
数组：有序的零个或者多个值。每个值可以为任意类型。序列表使用方括号[，]括起来。元素之间用逗号,分割。形如：[value, value]
对象：若干无序的“键-值对”(key-value pairs)，其中键只能是字符串[1]。建议但不强制要求对象中的键是独一无二的。对象以花括号{开始，并以}结束。键-值对之间使用逗号分隔。键与值之间用冒号:分割。
空值：值写为null

token（6种标点符号、字符串、数值、3种字面量）之间可以存在有限的空白符并被忽略。四个特定字符被认为是空白符：空格符、水平制表符、回车符、换行符。空白符不能出现在token内部（但空格符可以出现在字符串内部）。JSON标准不允许有字节序掩码，不提供注释的句法。 一个有效的JSON文档的根节点必须是一个对象或一个数组。

JSON交换时必须编码为UTF-8。[2]转义序列可以为：“\\”、“\"”、“\/”、“\b”、“\f”、“\n”、“\r”、“\t”，或Unicode16进制转义字符序列（\u后面跟随4位16进制数字）。对于不在基本多文种平面上的码位，必须用UTF-16代理对（surrogate pair）表示，例如对于Emoji字符——喜极而泣的表情（U+1F602 😂 face with tears of joy）在JSON中应表示为：
{ "face": "😂" }
// or
{ "face": "\uD83D\uDE02" }

注意：golang 的 []byte 比较特殊，会被编码为 base64 形式，需要特殊处理


------------
在 Go 中并不是所有的类型都能进行序列化：
	JSON object key 只支持 string
	Channel、complex、function 等 type 无法进行序列化
	数据中如果存在循环引用，则不能进行序列化，因为序列化时会进行递归
	Pointer 序列化之后是其指向的值或者是 nil
	只有 struct 中支持导出的 field 才能被 JSON package 序列化，即首字母大写的 field。
反序列化:
	`json:"field,string"`
	`json:"some_field,omitempty"`
	`json:"-"`
默认的 JSON 只支持以下几种 Go 类型：
	bool for JSON booleans
	float64 for JSON numbers
	string for JSON strings
	nil for JSON null
反序列化对 slice、map、pointer 的处理:
如果我们序列化之前不知道其数据格式，我们可以使用 interface{} 来存储我们的 decode 之后的数据：
	var f interface{}
	err := json.Unmarshal(b, &f)
	key 是 string，value 是存储在 interface{} 内的。想要获得 f 中的数据，我们首先需要进行 type assertion，
然后通过 range 迭代获得 f 中所有的 key ：
		m := f.(map[string]interface{})
		for k, v := range m {
			switch vv := v.(type) {
			case string:
				fmt.Println(k, "is string", vv)
			case float64:
				fmt.Println(k, "is float64", vv)
			case []interface{}:
				fmt.Println(k, "is an array:")
				for i, u := range vv {
					fmt.Println(i, u)
				}
			default:
				fmt.Println(k, "is of a type I don't know how to handle")
			}
		}
Stream JSON:
	除了 marshal 和 unmarshal 函数，Go 还提供了 Decoder 和 Encoder 对 stream JSON 进行处理，常见 request
中的 Body、文件等

嵌入式 struct 的序列化:
	Go 支持对 nested struct 进行序列化和反序列化:
自定义序列化函数:
	Go JSON package 中定了两个 Interface Marshaler 和 Unmarshaler ，实现这两个 Interface 可以让你定义的
type 支持序列化操作。

-----

1. 可以使用缓存池 sync.Pool(不合适，没有合适的回收时机，，，)， 或者自己实现生成池
2. 用 string 替换[]byte，因为前者可以复用 string 给输出端


Dec	Hex	Binary  	HTML	Char	Description
0 	00	00000000	&#0;	NUL 	Null
1 	01	00000001	&#1;	SOH 	Start of Header
2 	02	00000010	&#2;	STX 	Start of Text
3 	03	00000011	&#3;	ETX 	End of Text
4 	04	00000100	&#4;	EOT 	End of Transmission
5 	05	00000101	&#5;	ENQ 	Enquiry
6 	06	00000110	&#6;	ACK 	Acknowledge
7 	07	00000111	&#7;	BEL 	Bell
8 	08	00001000	&#8;	BS  	Backspace
9 	09	00001001	&#9;	HT  	Horizontal Tab
10	0A	00001010	&#10;	LF  	Line Feed
11	0B	00001011	&#11;	VT  	Vertical Tab
12	0C	00001100	&#12;	FF  	Form Feed
13	0D	00001101	&#13;	CR  	Carriage Return
14	0E	00001110	&#14;	SO  	Shift Out
15	0F	00001111	&#15;	SI  	Shift In
16	10	00010000	&#16;	DLE 	Data Link Escape
17	11	00010001	&#17;	DC1 	Device Control 1
18	12	00010010	&#18;	DC2 	Device Control 2
19	13	00010011	&#19;	DC3 	Device Control 3
20	14	00010100	&#20;	DC4 	Device Control 4
21	15	00010101	&#21;	NAK 	Negative Acknowledge
22	16	00010110	&#22;	SYN 	Synchronize
23	17	00010111	&#23;	ETB 	End of Transmission Block
24	18	00011000	&#24;	CAN 	Cancel
25	19	00011001	&#25;	EM  	End of Medium
26	1A	00011010	&#26;	SUB 	Substitute
27	1B	00011011	&#27;	ESC 	Escape
28	1C	00011100	&#28;	FS  	File Separator
29	1D	00011101	&#29;	GS  	Group Separator
30	1E	00011110	&#30;	RS  	Record Separator
31	1F	00011111	&#31;	US  	Unit Separator
32	20	00100000	&#32;	space	Space
33	21	00100001	&#33;	!   	exclamation mark
34	22	00100010	&#34;	"   	double quote
35	23	00100011	&#35;	#   	number
36	24	00100100	&#36;	$   	dollar
37	25	00100101	&#37;	%   	percent
38	26	00100110	&#38;	&   	ampersand
39	27	00100111	&#39;	'   	single quote
40	28	00101000	&#40;	(   	left parenthesis
41	29	00101001	&#41;	)   	right parenthesis
42	2A	00101010	&#42;	*   	asterisk
43	2B	00101011	&#43;	+   	plus
44	2C	00101100	&#44;	,   	comma
45	2D	00101101	&#45;	-   	minus
46	2E	00101110	&#46;	.   	period
47	2F	00101111	&#47;	/   	slash
48	30	00110000	&#48;	0   	zero
49	31	00110001	&#49;	1   	one
50	32	00110010	&#50;	2   	two
51	33	00110011	&#51;	3   	three
52	34	00110100	&#52;	4   	four
53	35	00110101	&#53;	5   	five
54	36	00110110	&#54;	6   	six
55	37	00110111	&#55;	7   	seven
56	38	00111000	&#56;	8   	eight
57	39	00111001	&#57;	9   	nine
58	3A	00111010	&#58;	:   	colon
59	3B	00111011	&#59;	;   	semicolon
60	3C	00111100	&#60;	<   	less than
61	3D	00111101	&#61;	=   	equality sign
62	3E	00111110	&#62;	>   	greater than
63	3F	00111111	&#63;	?   	question mark
64	40	01000000	&#64;	@   	at sign
65	41	01000001	&#65;	A	 
66	42	01000010	&#66;	B	 
67	43	01000011	&#67;	C	 
68	44	01000100	&#68;	D	 
69	45	01000101	&#69;	E	 
70	46	01000110	&#70;	F	 
71	47	01000111	&#71;	G	 
72	48	01001000	&#72;	H	 
73	49	01001001	&#73;	I	 
74	4A	01001010	&#74;	J	 
75	4B	01001011	&#75;	K	 
76	4C	01001100	&#76;	L	 
77	4D	01001101	&#77;	M	 
78	4E	01001110	&#78;	N	 
79	4F	01001111	&#79;	O	 
80	50	01010000	&#80;	P	 
81	51	01010001	&#81;	Q	 
82	52	01010010	&#82;	R	 
83	53	01010011	&#83;	S	 
84	54	01010100	&#84;	T	 
85	55	01010101	&#85;	U	 
86	56	01010110	&#86;	V	 
87	57	01010111	&#87;	W	 
88	58	01011000	&#88;	X	 
89	59	01011001	&#89;	Y	 
90	5A	01011010	&#90;	Z	 
91	5B	01011011	&#91;	[   	left square bracket
92	5C	01011100	&#92;	\   	backslash
93	5D	01011101	&#93;	]   	right square bracket
94	5E	01011110	&#94;	^   	caret / circumflex
95	5F	01011111	&#95;	_   	underscore
96	60	01100000	&#96;	`   	grave / accent
97	61	01100001	&#97;	a	 
98	62	01100010	&#98;	b	 
99	63	01100011	&#99;	c	 
100	64	01100100	&#100;	d	 
101	65	01100101	&#101;	e	 
102	66	01100110	&#102;	f	 
103	67	01100111	&#103;	g	 
104	68	01101000	&#104;	h	 
105	69	01101001	&#105;	i	 
106	6A	01101010	&#106;	j	 
107	6B	01101011	&#107;	k	 
108	6C	01101100	&#108;	l	 
109	6D	01101101	&#109;	m	 
110	6E	01101110	&#110;	n	 
111	6F	01101111	&#111;	o	 
112	70	01110000	&#112	p	 
113	71	01110001	&#113;	q	 
114	72	01110010	&#114;	r	 
115	73	01110011	&#115;	s	 
116	74	01110100	&#116;	t	 
117	75	01110101	&#117;	u	 
118	76	01110110	&#118;	v	 
119	77	01110111	&#119;	w	 
120	78	01111000	&#120;	x	 
121	79	01111001	&#121;	y	 
122	7A	01111010	&#122;	z	 
123	7B	01111011	&#123;	{   	left curly bracket
124	7C	01111100	&#124;	|   	vertical bar
125	7D	01111101	&#125;	}   	right curly bracket
126	7E	01111110	&#126;	~   	tilde
127	7F	01111111	&#127;	DEL  	delete