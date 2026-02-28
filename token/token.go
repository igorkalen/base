package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	
	IDENT  = "IDENT"  
	INT    = "INT"    
	FLOAT  = "FLOAT"  
	STRING = "STRING" 

	
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	MOD      = "%"

	LT = "<"
	GT = ">"

	EQ     = "=="
	NOT_EQ = "!="
	LTE    = "<="
	GTE    = ">="

	
	AND         = "and"
	OR          = "or"
	NOT         = "not"
	BIT_AND     = "&"
	BIT_OR      = "|"
	BIT_XOR     = "^"
	BIT_NOT     = "~"
	LEFT_SHIFT  = "<<"
	RIGHT_SHIFT = ">>"

	
	QUESTION = "?"
	COLON    = ":"

	
	COMMA     = ","
	SEMICOLON = ";"
	DOT       = "."

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	
	FUNCTION = "FUNCTION"
	LET      = "LET"
	GLOBAL   = "GLOBAL"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	WHILE    = "WHILE"
	FOR      = "FOR"
	FOREACH  = "FOREACH"
	IN       = "IN"
	IMPORT   = "IMPORT"
	AS       = "AS"
	TRY      = "TRY"
	CATCH    = "CATCH"
	THROW    = "THROW"
	ASYNC    = "ASYNC"
	SPAWN    = "SPAWN"
	SCHEDULE = "SCHEDULE"
)

var keywords = map[string]TokenType{
	"function": FUNCTION,
	"let":      LET,
	"global":   GLOBAL,
	"true":     TRUE,
	"false":    FALSE,
	"if":       IF,
	"else":     ELSE,
	"return":   RETURN,
	"while":    WHILE,
	"for":      FOR,
	"foreach":  FOREACH,
	"in":       IN,
	"import":   IMPORT,
	"as":       AS,
	"try":      TRY,
	"catch":    CATCH,
	"throw":    THROW,
	"async":    ASYNC,
	"spawn":    SPAWN,
	"schedule": SCHEDULE,
	"and":      AND,
	"or":       OR,
	"not":      NOT,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
