---- MODULE Exprs ----
EXTENDS Naturals
CONSTANT C
VARIABLE x
A == 1 + 2
B(y) == IF y \in {1, 2} THEN TRUE ELSE FALSE
Pair == <<C, x>>
Rec == [a |-> C, b |-> x]
SetMap == {z * 2 : z \in {1, 2}}
Fn == [z \in {1, 2} |-> z + C]
Let == LET D == C IN D
All == \A z \in {1, 2} : z >= 1
Choice == CHOOSE z \in {1, 2} : z = 1
CaseOp == CASE C = 1 -> TRUE [] OTHER -> FALSE
Act == [x' = x + 1]_x
====
