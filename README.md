tlago: TLA+ tools ported to Go
==============================

# repos with corpora/example TLA+ specs to check parsing/processing on

These are vendored under test_vectors/ now

There are several open-source repositories and standardized corpora that serve as test suites for TLA+ parsers.

Because TLA+ features tricky syntactic edge cases (such as whitespace-aligned conjunction/disjunction lists and complex operator precedence), testing against real-world specifications is essential.

Here are the best collections of TLA+ specs to test your parser against, ranked by utility:

---

### 1. `tlaplus/Examples` (The Standard Corpus)

This is the official, community-maintained repository of TLA+ specifications managed by the TLA+ Foundation. It contains hundreds of `.tla` files ranging from simple beginner puzzles to full industrial protocols (like Paxos and Raft).

* **Key specs to include in your parser suite:**
* `DieHard.tla` / `NQueens.tla`: Simple, clean syntax for initial sanity checks.
* `Paxos.tla` / `Raft.tla`: Large, complex distributed systems specifications with heavy operator nesting.
* `Specifying Systems` folder: The exact specifications featured in Leslie Lamport’s foundational textbook *Specifying Systems*.


* **Repository:** [github.com/tlaplus/Examples](https://github.com/tlaplus/Examples?utm_source=gemini)

### 2. Tree-Sitter TLA+ Corpus (Best for Parsing Edge Cases)

If you want focused, bite-sized test cases specifically tailored for grammar validation, check out the test suite written for `tree-sitter-tlaplus`.

* **Why it's useful:** Tree-sitter repositories organize tests into a structured `corpus/` folder containing hundreds of target syntax snippets paired with expected Abstract Syntax Trees (ASTs).
* **What it tests:** Highly specific parsing boundaries—such as nested alignment-based conjunction (`/\`) and disjunction (`\/`) lists, record constructs, tuple syntax, and operator fixity/precedence rules.
* **Repository:** [github.com/tlaplus-community/tree-sitter-tlaplus](https://github.com/tlaplus-community/tree-sitter-tlaplus?utm_source=gemini)


## Key Syntactic Stress-Tests for Your Parser

When testing your parser against these specifications, ensure your test suite exercises these specific TLA+ grammar features:

1. **Junction List Scoping:**
```tla
\* Ensure your parser respects column alignment to build the tree correctly
\/ A
\/ /\ B
   /\ C
\/ D

```

2. **Module Extensions & Instances:**
`EXTENDS Naturals, Sequences, TLC` and `INSTANCE ModuleName WITH x <- y`.

3. **Bounded & Unbounded Quantifiers:**
`\E x \in S : P(x)` vs. `\A x, y : P(x, y)`.

4. **Custom Infix & Prefix Operators:**
Operators like `x \div y`, `x ++ y`, or `x \cap y` with custom TLA+ precedence rules.

5. **PlusCal Comments:**
Many specs embed PlusCal algorithms inside `(* --algorithm ... *)` blocks. Your lexer/parser must either ignore these block comments or handle them cleanly without throwing syntax errors.

