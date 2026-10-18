----------------------------- MODULE Counter -----------------------------
EXTENDS Naturals

CONSTANT Limit
VARIABLE n

Init == n = 0
Next == n' = IF n < Limit THEN n + 1 ELSE 0

TypeOK == n \in 0..Limit
BelowLimit == n < Limit
=============================================================================
