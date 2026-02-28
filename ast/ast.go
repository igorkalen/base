package ast

import (
	"base/token"
	"bytes"
	"strings"
)


type Node interface {
	TokenLiteral() string
	String() string
}


type Statement interface {
	Node
	statementNode()
}


type Expression interface {
	Node
	expressionNode()
}


type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}


type Identifier struct {
	Token token.Token 
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }


type LetStatement struct {
	Token token.Token 
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}


type GlobalStatement struct {
	Token token.Token 
	Name  *Identifier
	Value Expression
}

func (gs *GlobalStatement) statementNode()       {}
func (gs *GlobalStatement) TokenLiteral() string { return gs.Token.Literal }
func (gs *GlobalStatement) String() string {
	var out bytes.Buffer

	out.WriteString(gs.TokenLiteral() + " ")
	out.WriteString(gs.Name.String())
	out.WriteString(" = ")

	if gs.Value != nil {
		out.WriteString(gs.Value.String())
	}

	out.WriteString(";")

	return out.String()
}


type AssignStatement struct {
	Token token.Token 
	Name  *Identifier
	Value Expression
}

func (as *AssignStatement) statementNode()       {}
func (as *AssignStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignStatement) String() string {
	var out bytes.Buffer

	out.WriteString(as.Name.String())
	out.WriteString(" = ")

	if as.Value != nil {
		out.WriteString(as.Value.String())
	}

	out.WriteString(";")

	return out.String()
}


type ReturnStatement struct {
	Token       token.Token 
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}


type ExpressionStatement struct {
	Token      token.Token 
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}


type BlockStatement struct {
	Token      token.Token 
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}


type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }


type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }


type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }


type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }


type PrefixExpression struct {
	Token    token.Token 
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}


type InfixExpression struct {
	Token    token.Token 
	Left     Expression
	Operator string
	Right    Expression
}

func (oe *InfixExpression) expressionNode()      {}
func (oe *InfixExpression) TokenLiteral() string { return oe.Token.Literal }
func (oe *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(oe.Left.String())
	out.WriteString(" " + oe.Operator + " ")
	out.WriteString(oe.Right.String())
	out.WriteString(")")

	return out.String()
}


type IfExpression struct {
	Token       token.Token 
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}


type FunctionLiteral struct {
	Token      token.Token 
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}


type CallExpression struct {
	Token     token.Token 
	Function  Expression  
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}


type PropertyAccessExpression struct {
	Token token.Token 
	Left  Expression
	Right *Identifier
}

func (pa *PropertyAccessExpression) expressionNode()      {}
func (pa *PropertyAccessExpression) TokenLiteral() string { return pa.Token.Literal }
func (pa *PropertyAccessExpression) String() string {
	return pa.Left.String() + "." + pa.Right.String()
}


type IndexExpression struct {
	Token token.Token 
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}


type WhileExpression struct {
	Token     token.Token 
	Condition Expression
	Body      *BlockStatement
}

func (we *WhileExpression) expressionNode()      {}
func (we *WhileExpression) TokenLiteral() string { return we.Token.Literal }
func (we *WhileExpression) String() string {
	var out bytes.Buffer

	out.WriteString("while")
	out.WriteString(we.Condition.String())
	out.WriteString(" ")
	out.WriteString(we.Body.String())

	return out.String()
}


type ForExpression struct {
	Token       token.Token 
	Initializer Statement
	Condition   Expression
	Increment   Statement
	Body        *BlockStatement
}

func (fe *ForExpression) expressionNode()      {}
func (fe *ForExpression) TokenLiteral() string { return fe.Token.Literal }
func (fe *ForExpression) String() string {
	var out bytes.Buffer

	out.WriteString("for (")

	if fe.Initializer != nil {
		out.WriteString(fe.Initializer.String())
	}
	out.WriteString("; ")

	if fe.Condition != nil {
		out.WriteString(fe.Condition.String())
	}
	out.WriteString("; ")

	if fe.Increment != nil {
		out.WriteString(fe.Increment.String())
	}
	out.WriteString(") ")

	out.WriteString(fe.Body.String())

	return out.String()
}


type ForEachExpression struct {
	Token    token.Token 
	ValueVar string      
	KeyVar   string      
	Iterable Expression  
	Body     *BlockStatement
}

func (fee *ForEachExpression) expressionNode()      {}
func (fee *ForEachExpression) TokenLiteral() string { return fee.Token.Literal }
func (fee *ForEachExpression) String() string {
	var out bytes.Buffer

	out.WriteString("foreach (")
	if fee.KeyVar != "" {
		out.WriteString(fee.KeyVar)
		out.WriteString(", ")
	}
	out.WriteString(fee.ValueVar)
	out.WriteString(" in ")
	out.WriteString(fee.Iterable.String())
	out.WriteString(") ")
	out.WriteString(fee.Body.String())

	return out.String()
}


type ArrayLiteral struct {
	Token    token.Token 
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}


type HashLiteral struct {
	Token token.Token 
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}


type TryCatchExpression struct {
	Token     token.Token 
	TryBody   *BlockStatement
	CatchVar  string 
	CatchBody *BlockStatement
}

func (tce *TryCatchExpression) expressionNode()      {}
func (tce *TryCatchExpression) TokenLiteral() string { return tce.Token.Literal }
func (tce *TryCatchExpression) String() string {
	var out bytes.Buffer

	out.WriteString("try ")
	out.WriteString(tce.TryBody.String())
	out.WriteString(" catch (")
	out.WriteString(tce.CatchVar)
	out.WriteString(") ")
	out.WriteString(tce.CatchBody.String())

	return out.String()
}


type ThrowStatement struct {
	Token token.Token 
	Value Expression
}

func (ts *ThrowStatement) statementNode()       {}
func (ts *ThrowStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *ThrowStatement) String() string {
	var out bytes.Buffer

	out.WriteString("throw ")
	if ts.Value != nil {
		out.WriteString(ts.Value.String())
	}
	out.WriteString(";")

	return out.String()
}


type ImportStatement struct {
	Token token.Token 
	Path  string
	Alias string
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	var out bytes.Buffer

	out.WriteString("import ")
	out.WriteString("\"" + is.Path + "\"")
	out.WriteString(" as ")
	out.WriteString(is.Alias)
	out.WriteString(";")

	return out.String()
}


type SpawnStatement struct {
	Token token.Token 
	Call  *CallExpression
}

func (ss *SpawnStatement) statementNode()       {}
func (ss *SpawnStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SpawnStatement) String() string {
	var out bytes.Buffer
	out.WriteString("spawn ")
	out.WriteString(ss.Call.String())
	out.WriteString(";")
	return out.String()
}


type TernaryExpression struct {
	Token       token.Token 
	Condition   Expression
	Consequence Expression
	Alternative Expression
}

func (te *TernaryExpression) expressionNode()      {}
func (te *TernaryExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TernaryExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(te.Condition.String())
	out.WriteString(" ? ")
	out.WriteString(te.Consequence.String())
	out.WriteString(" : ")
	out.WriteString(te.Alternative.String())
	out.WriteString(")")

	return out.String()
}
